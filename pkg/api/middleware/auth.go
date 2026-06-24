package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"go.uber.org/zap"
)

const (
	authorizationHeaderKey string = "Authorization"
	authorizationType      string = "Bearer"
)

// Get User Auth middleware
func (c *middleware) AuthenticateUser() gin.HandlerFunc {
	return c.authorize(token.User)
	// return c.middlewareUsingCookie(token.User)
}

// Get Admin Auth middleware
func (c *middleware) AuthenticateAdmin() gin.HandlerFunc {
	return c.authorize(token.Admin)
	// return c.middlewareUsingCookie(token.Admin)
}

// authorize request on request header using user type
func (c *middleware) authorize(tokenUser token.UserType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationValues := ctx.GetHeader(authorizationHeaderKey)
		authFields := strings.Fields(authorizationValues)

		if len(authFields) < 2 {
			err := errors.New("authorization token not provided properly with prefix of Bearer")
			applogger.SecurityEvent(ctx, applogger.EventSecurityAuthFailed,
				zap.String("reason", "missing_bearer_token"),
				zap.String("token_user_type", string(tokenUser)),
			).Warn("auth rejected: missing bearer token")
			response.ErrorResponse(ctx, http.StatusUnauthorized, "Failed to authorize request", err, nil)
			ctx.Abort()
			return
		}

		authType := authFields[0]
		accessToken := authFields[1]

		if !strings.EqualFold(authType, authorizationType) {
			err := errors.New("invalid authorization type")
			applogger.SecurityEvent(ctx, applogger.EventSecurityAuthFailed,
				zap.String("reason", "invalid_auth_type"),
				zap.String("got", authType),
			).Warn("auth rejected: invalid authorization type")
			response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized user", err, nil)
			ctx.Abort()
			return
		}

		tokenVerifyReq := token.VerifyTokenRequest{
			TokenString: accessToken,
			UsedFor:     tokenUser,
		}

		verifyRes, err := c.tokenService.VerifyToken(tokenVerifyReq)

		// if initial verify failed due to invalid token and we were checking for a user,
		// try verifying as admin so that admins can access user endpoints too.
		if err != nil && errors.Is(err, token.ErrInvalidToken) && tokenUser == token.User {
			altReq := token.VerifyTokenRequest{
				TokenString: accessToken,
				UsedFor:     token.Admin,
			}
			altRes, altErr := c.tokenService.VerifyToken(altReq)
			if altErr == nil {
				verifyRes = altRes
				err = nil
			}
		}

		if err != nil {
			applogger.SecurityEvent(ctx, applogger.EventSecurityTokenInvalid,
				zap.String("reason", "token_verification_failed"),
				zap.String("token_user_type", string(tokenUser)),
				zap.Error(err),
			).Warn("auth rejected: invalid token")
			response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized user", err, nil)
			ctx.Abort()
			return
		}

		ctx.Set("userId", verifyRes.UserID)
	}
}
