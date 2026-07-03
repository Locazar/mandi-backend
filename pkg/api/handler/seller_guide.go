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
	guideVideoNamespace    = "guide_video"
	trainingVideoNamespace = "training_video"
)

// SellerGuideHandler serves seller onboarding guide data and manages guide/training videos.
type SellerGuideHandler struct {
	cloudService cloud.CloudService
}

func NewSellerGuideHandler(cloudService cloud.CloudService) *SellerGuideHandler {
	return &SellerGuideHandler{cloudService: cloudService}
}

// ── Public ────────────────────────────────────────────────────────────────────

// GetShopPhotoTips GET /api/seller-guide/shop-photo-tips?video=<filename>
func (h *SellerGuideHandler) GetShopPhotoTips(ctx *gin.Context) {
	videoName := ctx.DefaultQuery("video", "video.mp4")
	videoURL := h.cloudService.PublicURL(guideVideoNamespace + "/" + videoName)
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

// GetPublicGuideVideos GET /api/seller-guide/videos — all guide and training videos, no auth (used by seller app)
func (h *SellerGuideHandler) GetPublicGuideVideos(ctx *gin.Context) {
	videos := make([]map[string]interface{}, 0)
	for _, namespace := range []string{guideVideoNamespace, trainingVideoNamespace} {
		keys, err := h.cloudService.ListObjects(ctx, namespace+"/")
		if err != nil {
			response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list videos", err, nil)
			return
		}
		for _, key := range keys {
			name := filepath.Base(key)
			if name == "" || name == "." || strings.HasSuffix(key, "/") {
				continue
			}
			videos = append(videos, map[string]interface{}{
				"name":      name,
				"key":       key,
				"type":      namespace,
				"video_url": h.cloudService.PublicURL(key),
			})
		}
	}
	response.SuccessResponse(ctx, http.StatusOK, "Videos retrieved", videos)
}

// GetCategories GET /api/seller-guide/categories
func (h *SellerGuideHandler) GetCategories(ctx *gin.Context) {
	categories := []map[string]interface{}{
		{"id": "1", "name": "Getting Started", "description": "Learn the basics of selling on Localzar"},
		{"id": "2", "name": "Product Management", "description": "How to add and manage your products"},
		{"id": "3", "name": "Orders & Fulfillment", "description": "Managing orders and delivery"},
		{"id": "4", "name": "Promotions & Offers", "description": "Running promotions and creating offers"},
		{"id": "5", "name": "Account & Settings", "description": "Account management and seller settings"},
	}
	response.SuccessResponse(ctx, http.StatusOK, "Categories retrieved", map[string]interface{}{
		"categories": categories,
	})
}

// ── Admin — Guide Videos ──────────────────────────────────────────────────────

// ListGuideVideos GET /api/admin/guide-videos
func (h *SellerGuideHandler) GetGuideVideos(ctx *gin.Context) {
	h.listVideos(ctx, guideVideoNamespace)
}

// UploadGuideVideo POST /api/admin/guide-videos  (form: video file, name?)
func (h *SellerGuideHandler) UploadGuideVideo(ctx *gin.Context) {
	h.saveVideo(ctx, guideVideoNamespace, http.StatusCreated, "Guide video uploaded")
}

// ReplaceGuideVideo PUT /api/admin/guide-videos  (form: video file, name?)
func (h *SellerGuideHandler) ReplaceGuideVideo(ctx *gin.Context) {
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

func (h *SellerGuideHandler) listVideos(ctx *gin.Context, namespace string) {
	keys, err := h.cloudService.ListObjects(ctx, namespace+"/")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list videos", err, nil)
		return
	}
	videos := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		name := filepath.Base(key)
		if name == "" || name == "." || strings.HasSuffix(key, "/") {
			continue
		}
		videos = append(videos, map[string]interface{}{
			"name":      name,
			"key":       key,
			"video_url": h.cloudService.PublicURL(key),
		})
	}
	response.SuccessResponse(ctx, http.StatusOK, "Videos retrieved", videos)
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
