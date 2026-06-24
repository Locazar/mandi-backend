package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// TraceIDHeader is propagated on every HTTP response so clients can
	// correlate their requests with backend log entries.
	TraceIDHeader = "X-Trace-ID"

	// ContextTraceID is the gin.Context key under which the trace ID is stored.
	// Other handlers and usecases can read it with ctx.GetString(logger.ContextTraceID).
	ContextTraceID = "trace_id"
)

// RequestLogger returns a Gin middleware that:
//   - Generates (or inherits) a trace_id per request
//   - Emits a structured JSON log line after the handler chain completes
//   - Tags every line with HTTP method, path, status, latency, client IP, user agent
//   - Writes the trace_id into the X-Trace-ID response header
//
// It replaces gin.Logger() and should be the first middleware on the engine.
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Honour an upstream trace ID (e.g. from an API gateway or load balancer).
		traceID := ctx.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		ctx.Set(ContextTraceID, traceID)
		ctx.Header(TraceIDHeader, traceID)

		// Let the handler chain run.
		ctx.Next()

		latency := time.Since(start)
		status := ctx.Writer.Status()
		path := ctx.FullPath()
		if path == "" {
			path = ctx.Request.URL.Path // no-route hit
		}

		fields := []zap.Field{
			zap.String("event", "api.request"),
			zap.String("trace_id", traceID),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("raw_path", ctx.Request.URL.RawPath),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("user_agent", ctx.Request.UserAgent()),
			zap.Int("response_size", ctx.Writer.Size()),
		}

		// Attach authenticated user ID when present.
		if userID, exists := ctx.Get("userId"); exists {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Attach first error if the handler set one.
		if len(ctx.Errors) > 0 {
			fields = append(fields, zap.String("error", ctx.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		log := L().With(fields...)
		msg := http.StatusText(status)

		switch {
		case status >= http.StatusInternalServerError:
			log.Error(msg)
		case status >= http.StatusBadRequest:
			log.Warn(msg)
		default:
			log.Info(msg)
		}
	}
}

// TraceIDFrom returns the trace ID stored in the Gin context, or an empty
// string if the request middleware has not run.
func TraceIDFrom(ctx *gin.Context) string {
	return ctx.GetString(ContextTraceID)
}
