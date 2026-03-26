package response

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type OTPResponse struct {
	OtpID string `json:"otp_id"`
}

// SendOTPResponse returned after sending OTP (no OTP in response)
type SendOTPResponse struct {
	Message            string `json:"message"`
	Phone              string `json:"phone"`
	OTPValiditySeconds int    `json:"otp_validity_seconds"`
	ConsentMessage     string `json:"consent_message"`
}

// VerifyOTPUserDetails represents user info returned after OTP verify
type VerifyOTPUserDetails struct {
	ID    uint   `json:"id"`
	Phone string `json:"phone"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// VerifyOTPResponse returned after a successful OTP verification
type VerifyOTPResponse struct {
	AccessToken string               `json:"access_token"`
	TokenType   string               `json:"token_type"`
	ExpiresIn   int                  `json:"expires_in"`
	User        VerifyOTPUserDetails `json:"user"`
	IsNewUser   bool                 `json:"is_new_user"`
	ConsentInfo string               `json:"consent_info"`
}
