package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/rohit221990/mandi-backend/internal/config"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
