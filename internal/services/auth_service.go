package services

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rohit221990/mandi-backend/internal/config"
	"github.com/rohit221990/mandi-backend/internal/models"
	"github.com/rohit221990/mandi-backend/internal/repositories"
	"github.com/rohit221990/mandi-backend/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg         *config.Config
	db          *gorm.DB
	rdb         *redis.Client
	users       *repositories.UserRepository
	otps        *repositories.OTPRepository
	audit       *repositories.AuditRepo
	primarySMS  utils.SMSClient
	fallbackSMS utils.SMSClient
}

func NewAuthService(cfg *config.Config, db *gorm.DB, rdb *redis.Client) *AuthService {
	return &AuthService{
		cfg:   cfg,
		db:    db,
		rdb:   rdb,
		users: repositories.NewUserRepository(db),
		otps:  repositories.NewOTPRepository(db),
		audit: repositories.NewAuditRepo(db),
		// prefer Twilio as primary and fallback to MSG91
		primarySMS:  utils.NewTwilio(cfg.SMS.TwilioSID, cfg.SMS.TwilioAuth),
		fallbackSMS: utils.NewMSG91(cfg.SMS.MSG91Key),
	}
}

// Register a new user (email/phone). Password optional for phone-only registration.
func (s *AuthService) Register(ctx context.Context, email, phone, password string) (*models.User, error) {
	// basic sanitization
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)
	if email == "" && phone == "" {
		return nil, errors.New("email or phone required")
	}
	var passHash string
	if password != "" {
		h, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}
		passHash = h
	}

	u := &models.User{Email: email, Phone: phone, PasswordHash: passHash}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login with email and password
func (s *AuthService) LoginWithPassword(ctx context.Context, email, password, ip, ua string) (string, string, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if err := utils.CompareHashAndPassword(u.PasswordHash, password); err != nil {
		_ = s.audit.Create(ctx, &models.LoginAuditLog{UserID: &u.ID, EventType: models.EventLoginFail, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()})
		return "", "", errors.New("invalid credentials")
	}
	access, refresh, err := utils.GenerateTokenPair(u.ID, s.cfg.JWT.AccessSecret, s.cfg.JWT.RefreshSecret, s.cfg.JWT.AccessTTL, s.cfg.JWT.RefreshTTL)
	if err != nil {
		return "", "", err
	}
	_ = s.audit.Create(ctx, &models.LoginAuditLog{UserID: &u.ID, EventType: models.EventLoginSuccess, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()})
	return access, refresh, nil
}

// SendOTP generates, stores hashed OTP and sends using primary SMS with fallback.
func (s *AuthService) SendOTP(ctx context.Context, target, ip, ua string) error {
	// rate-limiting in Redis
	phoneKey := fmt.Sprintf("otp:phone:%s", target)
	ipKey := fmt.Sprintf("otp:ip:%s", ip)

	// max 3 per phone per hour
	if n, _ := s.rdb.Get(ctx, phoneKey).Int(); n >= 3 {
		return errors.New("otp limit reached for phone")
	}
	if n, _ := s.rdb.Get(ctx, ipKey).Int(); n >= 10 {
		return errors.New("otp limit reached for ip")
	}

	// cooldown check
	cooldownKey := fmt.Sprintf("otp:cooldown:%s", target)
	if ttl, _ := s.rdb.TTL(ctx, cooldownKey).Result(); ttl > 0 {
		return errors.New("please wait before requesting another OTP")
	}

	otp := generateOTP()
	hashed, err := utils.HashPassword(otp)
	if err != nil {
		return err
	}

	o := &models.OTPRequest{Target: target, OTPHash: hashed, ExpiresAt: time.Now().Add(5 * time.Minute), Attempts: 0, MaxAttempt: 3, Status: models.OTPStatusPending, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()}
	if err := s.otps.Create(ctx, o); err != nil {
		return err
	}

	// increment counters and set cooldown
	s.rdb.Incr(ctx, phoneKey)
	s.rdb.Expire(ctx, phoneKey, time.Hour)
	s.rdb.Incr(ctx, ipKey)
	s.rdb.Expire(ctx, ipKey, time.Hour)
	s.rdb.Set(ctx, cooldownKey, "1", time.Minute)

	// try sending
	msg := fmt.Sprintf("Your OTP is %s. It is valid for 5 minutes.", otp)
	if err := s.primarySMS.SendSMS(ctx, target, msg); err != nil {
		_ = s.fallbackSMS.SendSMS(ctx, target, msg)
	}

	_ = s.audit.Create(ctx, &models.LoginAuditLog{UserID: nil, EventType: models.EventOTPRequested, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()})
	return nil
}

// VerifyOTP verifies code, increments attempts, invalidates after success
func (s *AuthService) VerifyOTP(ctx context.Context, target, code, ip, ua string) (bool, error) {
	o, err := s.otps.FindLatest(ctx, target)
	if err != nil {
		return false, err
	}
	if time.Now().After(o.ExpiresAt) {
		return false, errors.New("otp expired")
	}
	if o.Status != models.OTPStatusPending {
		return false, errors.New("otp not pending")
	}
	if o.Attempts >= o.MaxAttempt {
		o.Status = models.OTPStatusFailed
		_ = s.otps.Update(ctx, o)
		return false, errors.New("max attempts reached")
	}

	if err := utils.CompareHashAndPassword(o.OTPHash, code); err != nil {
		o.Attempts += 1
		_ = s.otps.Update(ctx, o)
		_ = s.audit.Create(ctx, &models.LoginAuditLog{UserID: nil, EventType: models.EventLoginFail, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()})
		return false, errors.New("invalid otp")
	}

	o.Status = models.OTPStatusUsed
	_ = s.otps.Update(ctx, o)
	_ = s.audit.Create(ctx, &models.LoginAuditLog{UserID: nil, EventType: models.EventOTPVerified, IPAddress: ip, UserAgent: ua, CreatedAt: time.Now()})
	return true, nil
}

// SetPIN sets a bcrypt-hashed PIN after OTP verification
func (s *AuthService) SetPIN(ctx context.Context, phone, pin string) error {
	u, err := s.users.FindByPhone(ctx, phone)
	if err != nil {
		return err
	}
	h, err := utils.HashPassword(pin)
	if err != nil {
		return err
	}
	u.PINHash = h
	return s.users.Update(ctx, u)
}

// Refresh issues new tokens if refresh token valid and not revoked
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.ParseToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return "", "", err
	}
	// Check blacklist
	blkKey := fmt.Sprintf("bl:%s", refreshToken)
	if b, _ := s.rdb.Get(ctx, blkKey).Result(); b != "" {
		return "", "", errors.New("token revoked")
	}
	access, refresh, err := utils.GenerateTokenPair(claims.UserID, s.cfg.JWT.AccessSecret, s.cfg.JWT.RefreshSecret, s.cfg.JWT.AccessTTL, s.cfg.JWT.RefreshTTL)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// Logout revokes refresh token by adding to Redis blacklist
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Parse token and compute remaining TTL, then set blacklist key in Redis
	tok, err := utils.ParseToken(refreshToken, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return err
	}
	if tok == nil {
		return errors.New("invalid token")
	}
	ttl := time.Until(tok.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	blkKey := fmt.Sprintf("bl:%s", refreshToken)
	return s.rdb.Set(ctx, blkKey, "1", ttl).Err()
}

func generateOTP() string {
	// generate secure 6-digit code using base32
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	s := base32.StdEncoding.EncodeToString(b)
	s = strings.ToUpper(s)
	// take numeric digits
	nums := ""
	for _, r := range s {
		if len(nums) >= 6 {
			break
		}
		if r >= '0' && r <= '9' {
			nums += string(r)
		}
	}
	if len(nums) < 6 {
		nums = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return nums[:6]
}
