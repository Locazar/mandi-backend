package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/internal/config"
	"github.com/rohit221990/mandi-backend/internal/utils"
)

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth"})
			return
		}
		// expected: Bearer <token>
		var token string
		if n, _ := fmt.Sscanf(auth, "Bearer %s", &token); n != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth header"})
			return
		}
		_, err := utils.ParseToken(token, cfg.JWT.AccessSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}
