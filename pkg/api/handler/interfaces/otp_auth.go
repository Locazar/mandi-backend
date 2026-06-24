// OTPAuthRequest represents the request body for OTP authentication
package interfaces

import "github.com/gin-gonic/gin"


type OTPAuthRequestHandler interface {
	SellerSignUpOtpSend(ctx *gin.Context)
	SellerSignUpOtpVerify(ctx *gin.Context)
}