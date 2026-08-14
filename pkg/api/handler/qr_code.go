package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// QRCodeHandler serves dynamic QR redirects: admin CRUD + PNG rendering, plus
// the public /r/:code redirect that scanners hit.
type QRCodeHandler struct {
	useCase *usecase.QRCodeUseCase
	// publicBaseURL is the externally reachable API origin used for the QR image
	// URL and, by default, the redirect link (e.g. https://api.locazar.com). When
	// blank we fall back to the request's own scheme+host, so local/dev still
	// works without config.
	publicBaseURL string
	// redirectBaseURL is the branded origin the *scannable* short link is built on
	// (e.g. https://locazar.com) so a scan shows a clean domain instead of the API
	// host. That domain must route /r/:code to this API. When blank the short link
	// falls back to publicBaseURL / the request host. Only shortURL uses it.
	redirectBaseURL string
}

func NewQRCodeHandler(useCase *usecase.QRCodeUseCase, publicBaseURL, redirectBaseURL string) *QRCodeHandler {
	return &QRCodeHandler{
		useCase:         useCase,
		publicBaseURL:   strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
		redirectBaseURL: strings.TrimRight(strings.TrimSpace(redirectBaseURL), "/"),
	}
}

// ── Request DTOs ─────────────────────────────────────────────────────────────

type createQRRequest struct {
	Title     string  `json:"title"`
	TargetURL string  `json:"target_url"`
	IsActive  *bool   `json:"is_active"`
	ExpiresAt *string `json:"expires_at"` // RFC3339; omit/null = never expires
}

// ── Admin CRUD ───────────────────────────────────────────────────────────────

// CreateQRCode godoc
//
//	@Summary	Create a dynamic QR redirect (Admin)
//	@Security	BearerAuth
//	@Tags		QR Codes
//	@Router		/admin/qr-codes [post]
func (h *QRCodeHandler) CreateQRCode(ctx *gin.Context) {
	var req createQRRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	expiresAt, err := parseOptionalTime(req.ExpiresAt)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid expires_at (use RFC3339)", err, nil)
		return
	}

	created, err := h.useCase.Create(ctx.Request.Context(), usecase.CreateInput{
		Title:     req.Title,
		TargetURL: req.TargetURL,
		IsActive:  req.IsActive,
		ExpiresAt: expiresAt,
		CreatedBy: utils.GetUserIdFromContext(ctx),
	})
	if err != nil {
		h.writeUseCaseError(ctx, err)
		return
	}

	h.enrich(ctx, &created)
	response.SuccessResponse(ctx, http.StatusCreated, "QR code created", created)
}

// ListQRCodes godoc
//
//	@Summary	List QR redirects (Admin)
//	@Security	BearerAuth
//	@Tags		QR Codes
//	@Router		/admin/qr-codes [get]
func (h *QRCodeHandler) ListQRCodes(ctx *gin.Context) {
	codes, err := h.useCase.List(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list QR codes", err, nil)
		return
	}
	for i := range codes {
		h.enrich(ctx, &codes[i])
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", codes)
}

// GetQRCode returns one code plus its most recent scan events.
func (h *QRCodeHandler) GetQRCode(ctx *gin.Context) {
	qr, err := h.useCase.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "QR code not found", err, nil)
		return
	}
	h.enrich(ctx, &qr)

	limit, _ := strconv.Atoi(ctx.Query("events_limit"))
	events, err := h.useCase.ListScanEvents(ctx.Request.Context(), qr.ID, limit)
	if err != nil {
		// Stats are non-critical — return the code without them rather than 500.
		events = []domain.QRScanEvent{}
	}
	response.SuccessResponse(ctx, http.StatusOK, "ok", gin.H{
		"qr_code":      qr,
		"recent_scans": events,
	})
}

// UpdateQRCode repoints/relabels/toggles a code. This is the "dynamic" part.
func (h *QRCodeHandler) UpdateQRCode(ctx *gin.Context) {
	// Bind into a raw map first so we can tell "expires_at omitted" from
	// "expires_at: null" (the latter clears the expiry).
	var raw map[string]interface{}
	if err := ctx.ShouldBindJSON(&raw); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	in := usecase.UpdateInput{}
	if v, ok := raw["title"].(string); ok {
		in.Title = &v
	}
	if v, ok := raw["target_url"].(string); ok {
		in.TargetURL = &v
	}
	if v, ok := raw["is_active"].(bool); ok {
		in.IsActive = &v
	}
	if rawExp, present := raw["expires_at"]; present {
		switch v := rawExp.(type) {
		case nil:
			in.ClearExpiry = true
		case string:
			if strings.TrimSpace(v) == "" {
				in.ClearExpiry = true
			} else {
				t, err := time.Parse(time.RFC3339, v)
				if err != nil {
					response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid expires_at (use RFC3339)", err, nil)
					return
				}
				in.ExpiresAt = &t
			}
		default:
			response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid expires_at", errors.New("expected RFC3339 string or null"), nil)
			return
		}
	}

	updated, err := h.useCase.Update(ctx.Request.Context(), ctx.Param("id"), in)
	if err != nil {
		h.writeUseCaseError(ctx, err)
		return
	}
	h.enrich(ctx, &updated)
	response.SuccessResponse(ctx, http.StatusOK, "QR code updated", updated)
}

// DeleteQRCode soft-deletes a code; its short link then resolves to 404.
func (h *QRCodeHandler) DeleteQRCode(ctx *gin.Context) {
	if err := h.useCase.Delete(ctx.Request.Context(), ctx.Param("id")); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.ErrorResponse(ctx, http.StatusNotFound, "QR code not found", err, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete QR code", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "QR code deleted", nil)
}

// GetQRImage renders the QR PNG for a code (encodes its short link). Admin-only;
// the admin-portal fetches it as a blob to preview/download.
func (h *QRCodeHandler) GetQRImage(ctx *gin.Context) {
	qr, err := h.useCase.GetByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "QR code not found", err, nil)
		return
	}

	size := 512
	if q := ctx.Query("size"); q != "" {
		if n, convErr := strconv.Atoi(q); convErr == nil {
			size = n
		}
	}
	if size < 128 {
		size = 128
	}
	if size > 2000 {
		size = 2000
	}

	png, err := qrcode.Encode(h.shortURL(ctx, qr.ShortCode), qrcode.Medium, size)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to render QR image", err, nil)
		return
	}
	ctx.Header("Content-Disposition", "inline; filename=\"qr-"+qr.ShortCode+".png\"")
	ctx.Header("Cache-Control", "no-store")
	ctx.Data(http.StatusOK, "image/png", png)
}

// ── Public redirect ──────────────────────────────────────────────────────────

// Redirect is the endpoint the QR encodes (GET /r/:code). It resolves the code
// and 302s the scanner to the target. Scan logging is best-effort and never
// blocks or fails the redirect.
func (h *QRCodeHandler) Redirect(ctx *gin.Context) {
	code := ctx.Param("code")

	res, err := h.useCase.Resolve(ctx.Request.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrQRDisabled):
			h.renderGone(ctx, "This link has been disabled.")
		case errors.Is(err, usecase.ErrQRExpired):
			h.renderGone(ctx, "This link has expired.")
		default:
			h.renderNotFound(ctx)
		}
		return
	}

	// Record the scan best-effort. Use the background context so it isn't tied
	// to the response lifecycle, and swallow errors — the redirect must win.
	go func(id, ip, ua, ref string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.useCase.RecordScan(bgCtx, id, ip, ua, ref)
	}(res.QRCode.ID, clientIP(ctx), ctx.Request.UserAgent(), ctx.Request.Referer())

	ctx.Redirect(http.StatusFound, res.TargetURL)
}

// ── helpers ──────────────────────────────────────────────────────────────────

// enrich fills the computed ShortURL/QRImageURL so the portal can render and
// download without knowing server topology.
func (h *QRCodeHandler) enrich(ctx *gin.Context, qr *domain.QRCode) {
	qr.ShortURL = h.shortURL(ctx, qr.ShortCode)
	qr.QRImageURL = h.base(ctx) + "/api/admin/qr-codes/" + qr.ID + "/image.png"
}

// shortURL is what the QR actually encodes: the public redirect link, on the
// branded redirect origin when configured (else the API base).
func (h *QRCodeHandler) shortURL(ctx *gin.Context, code string) string {
	return h.redirectBase(ctx) + "/r/" + code
}

// redirectBase is the origin the scannable short link is built on. It prefers
// the branded QRRedirectBaseURL (e.g. https://locazar.com) and falls back to the
// API base (publicBaseURL / request host) when unset.
func (h *QRCodeHandler) redirectBase(ctx *gin.Context) string {
	if h.redirectBaseURL != "" {
		return h.redirectBaseURL
	}
	return h.base(ctx)
}

// base is the origin used to build both the short link and the image URL. It
// prefers the configured PublicBaseURL (set this to the public HTTPS origin in
// production) and only falls back to deriving it from the request when unset.
func (h *QRCodeHandler) base(ctx *gin.Context) string {
	if h.publicBaseURL != "" {
		return h.publicBaseURL
	}
	return h.requestBase(ctx)
}

// requestBase derives scheme://host from the incoming request, honouring a
// reverse-proxy's X-Forwarded-Proto. Behind a TLS-terminating proxy that does
// not forward that header the scheme can come out http, which is why setting
// PublicBaseURL is recommended in production.
func (h *QRCodeHandler) requestBase(ctx *gin.Context) string {
	scheme := "http"
	if proto := ctx.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if ctx.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + ctx.Request.Host
}

func (h *QRCodeHandler) writeUseCaseError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrQRNotFound):
		response.ErrorResponse(ctx, http.StatusNotFound, "QR code not found", err, nil)
	case errors.Is(err, usecase.ErrQRTargetRequired),
		errors.Is(err, usecase.ErrQRTargetScheme),
		errors.Is(err, usecase.ErrQRTitleTooLong),
		errors.Is(err, usecase.ErrQRNothingToUpdate):
		response.ErrorResponse(ctx, http.StatusBadRequest, err.Error(), err, nil)
	case errors.Is(err, usecase.ErrQRShortCodeGen):
		response.ErrorResponse(ctx, http.StatusConflict, "Could not allocate a unique short code, please retry", err, nil)
	default:
		response.ErrorResponse(ctx, http.StatusInternalServerError, "QR code operation failed", err, nil)
	}
}

func (h *QRCodeHandler) renderNotFound(ctx *gin.Context) {
	ctx.Data(http.StatusNotFound, "text/html; charset=utf-8",
		[]byte(qrMessagePage("Link not found", "This QR code is invalid or no longer exists.")))
}

func (h *QRCodeHandler) renderGone(ctx *gin.Context, msg string) {
	ctx.Data(http.StatusGone, "text/html; charset=utf-8",
		[]byte(qrMessagePage("Link unavailable", msg)))
}

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// clientIP prefers proxy-forwarded client addresses, falling back to gin's.
func clientIP(ctx *gin.Context) string {
	if xff := ctx.GetHeader("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	return ctx.ClientIP()
}

func qrMessagePage(title, body string) string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8">` +
		`<meta name="viewport" content="width=device-width,initial-scale=1">` +
		`<title>` + title + `</title></head>` +
		`<body style="font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;` +
		`display:flex;min-height:100vh;margin:0;align-items:center;justify-content:center;` +
		`background:#f8f9fa;color:#2d3748;text-align:center">` +
		`<div><h1 style="font-size:20px;margin:0 0 8px">` + title + `</h1>` +
		`<p style="margin:0;color:#718096">` + body + `</p></div></body></html>`
}
