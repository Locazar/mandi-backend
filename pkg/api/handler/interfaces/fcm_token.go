package interfaces

import "github.com/gin-gonic/gin"

type FcmTokenHandler interface {
	SaveFcmToken(c *gin.Context)
}
