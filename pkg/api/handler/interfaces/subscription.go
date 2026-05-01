package interfaces

import "github.com/gin-gonic/gin"

type SubscriptionHandler interface {
	GetSubscriptionStatus(ctx *gin.Context)
	StartTrial(ctx *gin.Context)
	GetPaidPlans(ctx *gin.Context)
}
