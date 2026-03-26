package otp

type OtpAuth interface {
	SentOtp(phoneNumber string) (string, error)
	VerifyOtp(phoneNumber string, code string) (valid bool, err error)

	SentOtpEmail(email string) (string, error)
	VerifyOtpEmail(email string, code string) (valid bool, err error)
}
