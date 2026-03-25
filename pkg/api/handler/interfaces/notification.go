package interfaces

import (
	"github.com/gin-gonic/gin"
)

type NotificationHandler interface {
	// Token management
	RegisterDeviceToken(ctx *gin.Context)
	UnregisterDeviceToken(ctx *gin.Context)

	// Notifications
	SaveNotification(ctx *gin.Context)
	GetNotificationsBy(ctx *gin.Context)
	MarkNotificationAsRead(ctx *gin.Context)
	SendPushNotification(ctx *gin.Context)

	// Backward compat alias
	GenerateFCMToken(ctx *gin.Context)
}
