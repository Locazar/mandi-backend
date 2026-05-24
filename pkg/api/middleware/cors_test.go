package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	t.Helper()

	server := gin.New()
	server.Use(CORSMiddleware())
	server.POST("/api/admin/auth/signin/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/admin/auth/signin/", nil)
	req.Header.Set("Origin", "http://localhost:4000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, "*", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET,POST,PUT,PATCH,DELETE,OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Origin, Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))
}
