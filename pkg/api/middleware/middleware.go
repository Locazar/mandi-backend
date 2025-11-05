package middleware

import (
	"github.com/gin-gonic/gin"
<<<<<<< HEAD
	"github.com/rohit221990/mandi-backend/pkg/service/token"
=======
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/service/token"
>>>>>>> b9ab446 (Initial commit)
)

type Middleware interface {
	AuthenticateUser() gin.HandlerFunc
	AuthenticateAdmin() gin.HandlerFunc
	TrimSpaces() gin.HandlerFunc
}

type middleware struct {
	tokenService token.TokenService
}

func NewMiddleware(tokenService token.TokenService) Middleware {
	return &middleware{
		tokenService: tokenService,
	}
}
