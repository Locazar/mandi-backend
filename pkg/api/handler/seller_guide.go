package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
)

const (
	guideVideoNamespace             = "guide_video"
	trainingVideoNamespace          = "training_video"
	productUploadGuideVideoNS       = "product_upload_guide"
	productUploadGuideDefaultFolder = "_default"
	// productUploadGuideContentFile is the companion text file (admin-written
	// tips/copy shown below the video in the seller app) saved alongside each
	// department's video, inside the same folder.
	productUploadGuideContentFile = "content.txt"
)

// SellerGuideHandler serves seller onboarding guide data and manages guide/training videos.
type SellerGuideHandler struct {
	cloudService cloud.CloudService
}

func NewSellerGuideHandler(cloudService cloud.CloudService) *SellerGuideHandler {
	return &SellerGuideHandler{cloudService: cloudService}
}

// ── Public ────────────────────────────────────────────────────────────────────

// GetShopPhotoTips GET /api/seller-guide/shop-photo-tips
// Returns the single guide video (first object in the guide_video namespace).
func (h *SellerGuideHandler) GetShopPhotoTips(ctx *gin.Context) {
	keys, err := h.cloudService.ListObjects(ctx, guideVideoNamespace+"/")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load guide video", err, nil)
		return
	}
	videoURL := ""
	for _, key := range keys {
		name := filepath.Base(key)
		if name == "" || name == "." || strings.HasSuffix(key, "/") {
			continue
		}
		videoURL = h.cloudService.PublicURL(key)
		break
	}
	if videoURL == "" {
		response.ErrorResponse(ctx, http.StatusNotFound, "No guide video uploaded", nil, nil)
		return
	}
	tutorials := []map[string]interface{}{
		{
			"id":          "shop_photo_tips_1",
			"title":       "How to take a perfect shop photo",
			"description": "Learn how to frame, light, and capture your storefront for the best first impression on customers.",
			"video_url":   videoURL,
			"duration":    "",
		},
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop photo tips retrieved", tutorials)
}

// GetPublicGuideVideos GET /api/seller-guide/guide-videos — guide videos (always one), no auth.
func (h *SellerGuideHandler) GetPublicGuideVideos(ctx *gin.Context) {
	h.listVideos(ctx, guideVideoNamespace)
}

// GetPublicTrainingVideos GET /api/seller-guide/training-videos — all training videos, no auth.
// Backs the seller app's Video Tutorials page.
func (h *SellerGuideHandler) GetPublicTrainingVideos(ctx *gin.Context) {
	h.listVideos(ctx, trainingVideoNamespace)
}

// GetCategories GET /api/seller-guide/categories
func (h *SellerGuideHandler) GetCategories(ctx *gin.Context) {
	categories := []map[string]interface{}{
		{"id": "1", "name": "Getting Started", "description": "Learn the basics of selling on Locazar"},
		{"id": "2", "name": "Product Management", "description": "How to add and manage your products"},
		{"id": "3", "name": "Orders & Fulfillment", "description": "Managing orders and delivery"},
		{"id": "4", "name": "Promotions & Offers", "description": "Running promotions and creating offers"},
		{"id": "5", "name": "Account & Settings", "description": "Account management and seller settings"},
	}
	response.SuccessResponse(ctx, http.StatusOK, "Categories retrieved", map[string]interface{}{
		"categories": categories,
	})
}

// GetPublicProductUploadGuideVideo GET /api/seller-guide/product-upload-guide-video?department=<name>
// Returns the single "how to upload a product" video for the given department,
// falling back to the default video if none was uploaded for that department.
func (h *SellerGuideHandler) GetPublicProductUploadGuideVideo(ctx *gin.Context) {
	department := ctx.Query("department")
	folder := slugifyDepartment(department)

	video, err := h.firstVideoInNamespace(ctx, productUploadGuideVideoNS+"/"+folder)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load guide video", err, nil)
		return
	}
	if video == nil {
		folder = productUploadGuideDefaultFolder
		video, err = h.firstVideoInNamespace(ctx, productUploadGuideVideoNS+"/"+folder)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load guide video", err, nil)
			return
		}
	}
	if video == nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "No product upload guide video uploaded", nil, nil)
		return
	}
	video["content"] = h.readProductUploadGuideContent(ctx, folder)
	response.SuccessResponse(ctx, http.StatusOK, "Product upload guide video retrieved", video)
}

// ── Admin — Product Upload Guide Videos ──────────────────────────────────────

// GetProductUploadGuideVideos GET /api/admin/product-upload-guide-videos?department=<name|"default">
func (h *SellerGuideHandler) GetProductUploadGuideVideos(ctx *gin.Context) {
	folder := productUploadGuideFolder(ctx.Query("department"))
	videos, err := h.listVideosData(ctx, productUploadGuideVideoNS+"/"+folder)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list videos", err, nil)
		return
	}
	content := h.readProductUploadGuideContent(ctx, folder)
	for _, v := range videos {
		v["content"] = content
	}
	response.SuccessResponse(ctx, http.StatusOK, "Videos retrieved", videos)
}

// UploadProductUploadGuideVideo POST /api/admin/product-upload-guide-videos
// (form: video file, department (or "default"), name?, content?)
// Each department folder holds exactly one video, so any existing one is removed first.
func (h *SellerGuideHandler) UploadProductUploadGuideVideo(ctx *gin.Context) {
	folder := productUploadGuideFolder(ctx.Query("department"))
	if folder == "" {
		folder = productUploadGuideFolder(ctx.PostForm("department"))
	}
	namespace := productUploadGuideVideoNS + "/" + folder
	if !h.clearNamespace(ctx, namespace) {
		return
	}
	if !h.saveProductUploadGuideContent(ctx, namespace) {
		return
	}
	h.saveVideo(ctx, namespace, http.StatusCreated, "Product upload guide video uploaded")
}

// ReplaceProductUploadGuideVideo PUT /api/admin/product-upload-guide-videos
// (form: video file, department (or "default"), name?, content?)
func (h *SellerGuideHandler) ReplaceProductUploadGuideVideo(ctx *gin.Context) {
	folder := productUploadGuideFolder(ctx.Query("department"))
	if folder == "" {
		folder = productUploadGuideFolder(ctx.PostForm("department"))
	}
	namespace := productUploadGuideVideoNS + "/" + folder
	if !h.clearNamespace(ctx, namespace) {
		return
	}
	if !h.saveProductUploadGuideContent(ctx, namespace) {
		return
	}
	h.saveVideo(ctx, namespace, http.StatusOK, "Product upload guide video replaced")
}

// DeleteProductUploadGuideVideo DELETE /api/admin/product-upload-guide-videos?department=<name|"default">&name=<filename>
func (h *SellerGuideHandler) DeleteProductUploadGuideVideo(ctx *gin.Context) {
	folder := productUploadGuideFolder(ctx.Query("department"))
	h.deleteVideo(ctx, productUploadGuideVideoNS+"/"+folder)
}

// ── Admin — Guide Videos ──────────────────────────────────────────────────────

// ListGuideVideos GET /api/admin/guide-videos
func (h *SellerGuideHandler) GetGuideVideos(ctx *gin.Context) {
	h.listVideos(ctx, guideVideoNamespace)
}

// UploadGuideVideo POST /api/admin/guide-videos  (form: video file, name?)
// The guide namespace holds exactly one video, so any existing ones are removed first.
func (h *SellerGuideHandler) UploadGuideVideo(ctx *gin.Context) {
	if !h.clearNamespace(ctx, guideVideoNamespace) {
		return
	}
	h.saveVideo(ctx, guideVideoNamespace, http.StatusCreated, "Guide video uploaded")
}

// ReplaceGuideVideo PUT /api/admin/guide-videos  (form: video file, name?)
func (h *SellerGuideHandler) ReplaceGuideVideo(ctx *gin.Context) {
	if !h.clearNamespace(ctx, guideVideoNamespace) {
		return
	}
	h.saveVideo(ctx, guideVideoNamespace, http.StatusOK, "Guide video replaced")
}

// DeleteGuideVideo DELETE /api/admin/guide-videos?name=<filename>
func (h *SellerGuideHandler) DeleteGuideVideo(ctx *gin.Context) {
	h.deleteVideo(ctx, guideVideoNamespace)
}

// ── Admin — Training Videos ───────────────────────────────────────────────────

// ListTrainingVideos GET /api/admin/training-videos
func (h *SellerGuideHandler) GetTrainingVideos(ctx *gin.Context) {
	h.listVideos(ctx, trainingVideoNamespace)
}

// UploadTrainingVideo POST /api/admin/training-videos  (form: video file, name?)
func (h *SellerGuideHandler) UploadTrainingVideo(ctx *gin.Context) {
	h.saveVideo(ctx, trainingVideoNamespace, http.StatusCreated, "Training video uploaded")
}

// ReplaceTrainingVideo PUT /api/admin/training-videos  (form: video file, name?)
func (h *SellerGuideHandler) ReplaceTrainingVideo(ctx *gin.Context) {
	h.saveVideo(ctx, trainingVideoNamespace, http.StatusOK, "Training video replaced")
}

// DeleteTrainingVideo DELETE /api/admin/training-videos?name=<filename>
func (h *SellerGuideHandler) DeleteTrainingVideo(ctx *gin.Context) {
	h.deleteVideo(ctx, trainingVideoNamespace)
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// slugifyDepartment normalizes a department name into a storage-safe folder
// name: lowercase, alphanumerics kept, everything else collapsed to "-".
func slugifyDepartment(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return productUploadGuideDefaultFolder
	}
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return productUploadGuideDefaultFolder
	}
	return slug
}

// productUploadGuideFolder resolves the storage folder for a department query
// param, treating "" and "default" as the shared fallback folder.
func productUploadGuideFolder(department string) string {
	if strings.EqualFold(strings.TrimSpace(department), "default") {
		return productUploadGuideDefaultFolder
	}
	return slugifyDepartment(department)
}

// firstVideoInNamespace returns the first video found under namespace, or nil
// (with no error) if the namespace has no videos.
func (h *SellerGuideHandler) firstVideoInNamespace(ctx *gin.Context, namespace string) (map[string]interface{}, error) {
	keys, err := h.cloudService.ListObjects(ctx, namespace+"/")
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		name := filepath.Base(key)
		if name == "" || name == "." || strings.HasSuffix(key, "/") || name == productUploadGuideContentFile {
			continue
		}
		return map[string]interface{}{
			"name":      name,
			"key":       key,
			"video_url": h.cloudService.PublicURL(key),
		}, nil
	}
	return nil, nil
}

// readProductUploadGuideContent reads the companion tips/copy text saved
// alongside a department's video. Returns "" if none was uploaded.
func (h *SellerGuideHandler) readProductUploadGuideContent(ctx *gin.Context, folder string) string {
	key := productUploadGuideVideoNS + "/" + folder + "/" + productUploadGuideContentFile
	data, err := h.cloudService.GetBytes(ctx, key)
	if err != nil {
		return ""
	}
	return string(data)
}

// saveProductUploadGuideContent saves the optional "content" form field as a
// text file alongside the video, so it survives independently of the video
// filename. Returns false (with an error response already written) on
// failure; a missing/empty "content" field is not an error.
func (h *SellerGuideHandler) saveProductUploadGuideContent(ctx *gin.Context, namespace string) bool {
	content := ctx.PostForm("content")
	if content == "" {
		return true
	}
	_, err := h.cloudService.SaveBytes(ctx, []byte(content), cloud.SaveOptions{
		Namespace:   namespace,
		Filename:    productUploadGuideContentFile,
		ContentType: "text/plain; charset=utf-8",
		Visibility:  cloud.VisibilityPublic,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save guide content", err, nil)
		return false
	}
	return true
}

// clearNamespace deletes every object under the namespace. Returns false if it
// failed and an error response was already written.
func (h *SellerGuideHandler) clearNamespace(ctx *gin.Context, namespace string) bool {
	keys, err := h.cloudService.ListObjects(ctx, namespace+"/")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list existing videos", err, nil)
		return false
	}
	for _, key := range keys {
		if strings.HasSuffix(key, "/") {
			continue
		}
		if err := h.cloudService.DeleteObject(ctx, key); err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to remove existing video", err, nil)
			return false
		}
	}
	return true
}

func (h *SellerGuideHandler) listVideos(ctx *gin.Context, namespace string) {
	videos, err := h.listVideosData(ctx, namespace)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list videos", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Videos retrieved", videos)
}

func (h *SellerGuideHandler) listVideosData(ctx *gin.Context, namespace string) ([]map[string]interface{}, error) {
	keys, err := h.cloudService.ListObjects(ctx, namespace+"/")
	if err != nil {
		return nil, err
	}
	videos := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		name := filepath.Base(key)
		if name == "" || name == "." || strings.HasSuffix(key, "/") || name == productUploadGuideContentFile {
			continue
		}
		videos = append(videos, map[string]interface{}{
			"name":      name,
			"key":       key,
			"video_url": h.cloudService.PublicURL(key),
		})
	}
	return videos, nil
}

func (h *SellerGuideHandler) saveVideo(ctx *gin.Context, namespace string, status int, msg string) {
	fh, err := ctx.FormFile("video")
	if err != nil || fh == nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "video file is required", err, nil)
		return
	}
	name := ctx.PostForm("name")
	if name == "" {
		name = fh.Filename
	}
	objectKey, err := h.cloudService.SaveFile(ctx, fh, cloud.SaveOptions{
		Namespace:  namespace,
		Filename:   name,
		Visibility: cloud.VisibilityPublic,
	})
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to upload video", err, nil)
		return
	}
	response.SuccessResponse(ctx, status, msg, map[string]interface{}{
		"name":      name,
		"key":       objectKey,
		"video_url": h.cloudService.PublicURL(objectKey),
	})
}

func (h *SellerGuideHandler) deleteVideo(ctx *gin.Context, namespace string) {
	name := ctx.Query("name")
	if name == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "name query param is required", nil, nil)
		return
	}
	objectKey := namespace + "/" + name
	if err := h.cloudService.DeleteObject(ctx, objectKey); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete video", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Video deleted", nil)
}
