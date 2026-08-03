package usecase

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

// Sentinel errors let the handler map failures to precise HTTP statuses without
// leaking internals.
var (
	ErrQRTargetRequired  = errors.New("target_url is required")
	ErrQRTargetScheme    = errors.New("target_url must be an absolute http(s) URL")
	ErrQRTitleTooLong    = errors.New("title must be 150 characters or fewer")
	ErrQRNotFound        = errors.New("qr code not found")
	ErrQRShortCodeGen    = errors.New("could not allocate a unique short code")
	ErrQRDisabled        = errors.New("qr code is disabled")
	ErrQRExpired         = errors.New("qr code has expired")
	ErrQRNothingToUpdate = errors.New("no updatable fields supplied")
)

// shortCodeAlphabet is base62 minus visually ambiguous characters, so codes are
// safe to print, read aloud, and embed in a low-density QR.
const shortCodeAlphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const (
	shortCodeLen    = 8
	shortCodeTries  = 6
	maxTitleLen     = 150
	scanEventBudget = 200
)

// QRCodeUseCase holds the business rules for dynamic QR redirects: validation,
// unique short-code allocation, and resolve-and-track on scan.
type QRCodeUseCase struct {
	repo repo.QRCodeRepository
}

func NewQRCodeUseCase(repo repo.QRCodeRepository) *QRCodeUseCase {
	return &QRCodeUseCase{repo: repo}
}

// CreateInput is the validated shape for creating a QR code.
type CreateInput struct {
	Title     string
	TargetURL string
	IsActive  *bool
	ExpiresAt *time.Time
	CreatedBy string
}

// UpdateInput carries only the fields the caller wants changed; nil means leave
// as-is. This keeps PATCH-style partial updates unambiguous.
type UpdateInput struct {
	Title     *string
	TargetURL *string
	IsActive  *bool
	ExpiresAt *time.Time // pointer-to-value set; see ClearExpiry to unset
	// ClearExpiry, when true, sets expires_at back to NULL regardless of
	// ExpiresAt. Lets a caller remove an expiry (which a nil pointer can't
	// express on its own).
	ClearExpiry bool
}

// normalizeTargetURL trims, validates and canonicalizes a redirect target.
// Only absolute http/https URLs are allowed — this is the guard against
// javascript:, data:, mailto: and other open-redirect abuse vectors.
func normalizeTargetURL(raw string) (string, error) {
	t := strings.TrimSpace(raw)
	if t == "" {
		return "", ErrQRTargetRequired
	}
	u, err := url.Parse(t)
	if err != nil {
		return "", ErrQRTargetScheme
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrQRTargetScheme
	}
	if u.Host == "" {
		return "", ErrQRTargetScheme
	}
	return u.String(), nil
}

func (u *QRCodeUseCase) Create(ctx context.Context, in CreateInput) (domain.QRCode, error) {
	title := strings.TrimSpace(in.Title)
	if len(title) > maxTitleLen {
		return domain.QRCode{}, ErrQRTitleTooLong
	}
	target, err := normalizeTargetURL(in.TargetURL)
	if err != nil {
		return domain.QRCode{}, err
	}
	if in.ExpiresAt != nil && in.ExpiresAt.Before(time.Now()) {
		return domain.QRCode{}, fmt.Errorf("expires_at must be in the future")
	}

	code, err := u.allocateShortCode(ctx)
	if err != nil {
		return domain.QRCode{}, err
	}

	qr := domain.QRCode{
		ShortCode: code,
		Title:     title,
		TargetURL: target,
		IsActive:  true,
		ExpiresAt: in.ExpiresAt,
		CreatedBy: in.CreatedBy,
	}
	if in.IsActive != nil {
		qr.IsActive = *in.IsActive
	}
	return u.repo.Create(ctx, qr)
}

func (u *QRCodeUseCase) List(ctx context.Context) ([]domain.QRCode, error) {
	return u.repo.List(ctx)
}

func (u *QRCodeUseCase) GetByID(ctx context.Context, id string) (domain.QRCode, error) {
	qr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.QRCode{}, ErrQRNotFound
	}
	return qr, nil
}

func (u *QRCodeUseCase) Update(ctx context.Context, id string, in UpdateInput) (domain.QRCode, error) {
	qr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return domain.QRCode{}, ErrQRNotFound
	}

	touched := false
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if len(t) > maxTitleLen {
			return domain.QRCode{}, ErrQRTitleTooLong
		}
		qr.Title = t
		touched = true
	}
	if in.TargetURL != nil {
		target, err := normalizeTargetURL(*in.TargetURL)
		if err != nil {
			return domain.QRCode{}, err
		}
		qr.TargetURL = target
		touched = true
	}
	if in.IsActive != nil {
		qr.IsActive = *in.IsActive
		touched = true
	}
	if in.ClearExpiry {
		qr.ExpiresAt = nil
		touched = true
	} else if in.ExpiresAt != nil {
		if in.ExpiresAt.Before(time.Now()) {
			return domain.QRCode{}, fmt.Errorf("expires_at must be in the future")
		}
		qr.ExpiresAt = in.ExpiresAt
		touched = true
	}
	if !touched {
		return domain.QRCode{}, ErrQRNothingToUpdate
	}
	return u.repo.Update(ctx, qr)
}

func (u *QRCodeUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.SoftDelete(ctx, id)
}

func (u *QRCodeUseCase) ListScanEvents(ctx context.Context, id string, limit int) ([]domain.QRScanEvent, error) {
	return u.repo.ListScanEvents(ctx, id, limit)
}

// ResolveResult is what the redirect handler needs to act: either a URL to send
// the scanner to, or a reason it can't.
type ResolveResult struct {
	QRCode    domain.QRCode
	TargetURL string
}

// Resolve looks up a short code and returns its target if it's live. It does
// NOT record the scan (that is a separate best-effort step) so a tracking
// failure can never change the redirect outcome.
func (u *QRCodeUseCase) Resolve(ctx context.Context, shortCode string) (ResolveResult, error) {
	code := strings.TrimSpace(shortCode)
	if code == "" {
		return ResolveResult{}, ErrQRNotFound
	}
	qr, err := u.repo.GetByShortCode(ctx, code)
	if err != nil {
		return ResolveResult{}, ErrQRNotFound
	}
	now := time.Now()
	if !qr.IsActive {
		return ResolveResult{QRCode: qr}, ErrQRDisabled
	}
	if qr.IsExpired(now) {
		return ResolveResult{QRCode: qr}, ErrQRExpired
	}
	return ResolveResult{QRCode: qr, TargetURL: qr.TargetURL}, nil
}

// RecordScan persists one scan event and bumps counters. Callers invoke it
// best-effort (errors logged, never surfaced to the scanner).
func (u *QRCodeUseCase) RecordScan(ctx context.Context, qrCodeID, ip, userAgent, referer string) error {
	return u.repo.RecordScan(ctx, domain.QRScanEvent{
		QRCodeID:  qrCodeID,
		IPAddress: truncate(ip, 64),
		UserAgent: truncate(userAgent, 1024),
		Referer:   truncate(referer, 1024),
	})
}

// allocateShortCode generates random codes until one is free, bounded by
// shortCodeTries so a pathological collision streak can't spin forever.
func (u *QRCodeUseCase) allocateShortCode(ctx context.Context) (string, error) {
	for i := 0; i < shortCodeTries; i++ {
		code, err := randomShortCode(shortCodeLen)
		if err != nil {
			return "", err
		}
		exists, err := u.repo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", ErrQRShortCodeGen
}

func randomShortCode(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = shortCodeAlphabet[int(b)%len(shortCodeAlphabet)]
	}
	return string(out), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
