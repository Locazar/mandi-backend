package request

type Login struct {
	Phone    string `json:"phone" binding:"omitempty,min=10,max=10"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"required,min=5,max=30"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"min=10"`
}

type OTPLoginEmail struct {
	Email string `json:"email" binding:"required,email"`
}

type RefreshSession struct {
	TokenID      string `json:"token_id"`
	UserID       uint   `json:"user_id"`
	UserType     string `json:"user_type"`
	RefreshToken string `json:"refresh_token"`
	ExpireAt     string `json:"expire_at"`
}

// VerifyOTPRequest is used to verify OTP sent to mobile
type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=10,max=15"`
	OTP   string `json:"otp" binding:"required,len=6"`
}
