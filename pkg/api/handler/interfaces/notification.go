package interfaces

import (
	"github.com/gin-gonic/gin"
)

type NotificationHandler interface {
	SaveNotification(ctx *gin.Context)
	GetNotificationsBy(ctx *gin.Context)
	MarkNotificationAsRead(ctx *gin.Context)
	GenerateFCMToken(ctx *gin.Context)
}
