package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/internal/config"
	"github.com/rohit221990/mandi-backend/internal/models"
	"github.com/rohit221990/mandi-backend/internal/services"
)

type AuthHandler struct {
	svc *services.AuthService
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) (*AuthHandler, error) {
	db, err := cfg.NewGorm()
	if err != nil {
		return nil, err
	}
	rdb := cfg.NewRedis()

	// auto-migrate models
	_ = db.AutoMigrate(&models.User{}, &models.OTPRequest{}, &models.LoginAuditLog{})

	svc := services.NewAuthService(cfg, db, rdb)
	return &AuthHandler{svc: svc, cfg: cfg}, nil
}

type registerReq struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,min=6"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var r registerReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.svc.Register(context.Background(), r.Email, r.Phone, r.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": u})
}

type loginReq struct {
	Email, Password string `json:"email"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var r loginReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	access, refresh, err := h.svc.LoginWithPassword(context.Background(), r.Email, r.Password, ip, ua)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access": access, "refresh": refresh})
}

type sendOTPReq struct {
	Target string `json:"target" binding:"required"`
}

func (h *AuthHandler) SendOTP(c *gin.Context) {
	var r sendOTPReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	if err := h.svc.SendOTP(context.Background(), r.Target, ip, ua); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "otp_sent"})
}

type verifyReq struct {
	Target, Code string `json:"target" binding:"required"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var r verifyReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ip := c.ClientIP()
	ua := c.Request.UserAgent()
	ok, err := h.svc.VerifyOTP(context.Background(), r.Target, r.Code, ip, ua)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"verified": ok})
}

type setPinReq struct {
	Phone, Pin string `json:"phone" binding:"required"`
}

func (h *AuthHandler) SetPIN(c *gin.Context) {
	var r setPinReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetPIN(context.Background(), r.Phone, r.Pin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pin_set"})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var r refreshReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	access, refresh, err := h.svc.Refresh(context.Background(), r.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access": access, "refresh": refresh})
}

type logoutReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var r logoutReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Logout(context.Background(), r.RefreshToken); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged_out"})
}
