package interfaces

import "github.com/gin-gonic/gin"

type SubscriptionPaymentHandler interface {
	CreateSubscriptionOrder(ctx *gin.Context)
	VerifySubscriptionPayment(ctx *gin.Context)
	HandlePaymentFailure(ctx *gin.Context)
	HandleWebhook(ctx *gin.Context)
}
