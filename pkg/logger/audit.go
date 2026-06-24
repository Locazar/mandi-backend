// audit.go exposes typed, domain-specific log helpers so that call sites never
// need to remember field names or which sub-logger to use.
//
// Every helper returns a *zap.Logger pre-loaded with the relevant fields.
// Callers finish the log entry with .Info / .Warn / .Error:
//
//	logger.ProductEvent(ctx, "product.created").Info("product created",
//	    zap.String("product_id", id))
package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ── event name constants ───────────────────────────────────────────────────────

const (
	// API
	EventAPIRequest = "api.request"

	// Security / auth
	EventSecurityAuthFailed      = "security.auth_failed"
	EventSecurityTokenInvalid    = "security.token_invalid"
	EventSecurityRateLimited     = "security.rate_limited"
	EventSecurityBruteForcBlock  = "security.brute_force_blocked"
	EventSecurityAdminAction     = "security.admin_action"
	EventSecurityOTPSent         = "security.otp_sent"
	EventSecurityOTPVerified     = "security.otp_verified"
	EventSecurityOTPFailed       = "security.otp_failed"
	EventSecurityLoginSuccess    = "security.login_success"
	EventSecurityLogout          = "security.logout"

	// Product lifecycle
	EventProductCreated = "product.created"
	EventProductUpdated = "product.updated"
	EventProductDeleted = "product.deleted"

	// Category / Department
	EventDepartmentCreated    = "catalog.department_created"
	EventDepartmentUpdated    = "catalog.department_updated"
	EventCategoryCreated      = "catalog.category_created"
	EventCategoryUpdated      = "catalog.category_updated"
	EventSubCategoryCreated   = "catalog.subcategory_created"
	EventSubCategoryUpdated   = "catalog.subcategory_updated"

	// Enquiry
	EventEnquiryCreated    = "enquiry.created"
	EventEnquiryAccepted   = "enquiry.accepted"
	EventEnquiryRejected   = "enquiry.rejected"
	EventEnquiryProcessed  = "enquiry.processed"

	// Payment
	EventPaymentInitiated = "payment.initiated"
	EventPaymentSuccess   = "payment.success"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"

	// Order
	EventOrderCreated   = "order.created"
	EventOrderUpdated   = "order.updated"
	EventOrderCancelled = "order.cancelled"
)

// ── generic helpers ───────────────────────────────────────────────────────────

// WithTrace returns the global logger enriched with trace_id extracted from the
// Gin context. Use this when you need a plain logger inside a handler.
func WithTrace(ctx *gin.Context) *zap.Logger {
	return L().With(zap.String("trace_id", TraceIDFrom(ctx)))
}

// ── domain event helpers ──────────────────────────────────────────────────────

// SecurityEvent returns the audit sub-logger pre-tagged for a security event.
// The returned logger writes to both app.log and audit.log.
//
//	logger.SecurityEvent(ctx, logger.EventSecurityAuthFailed,
//	    zap.String("reason", "invalid bearer token")).Warn("auth failed")
func SecurityEvent(ctx *gin.Context, event string, extra ...zap.Field) *zap.Logger {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("trace_id", TraceIDFrom(ctx)),
		zap.String("client_ip", ctx.ClientIP()),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	}
	fields = append(fields, extra...)
	return Audit().With(fields...)
}

// ProductEvent returns a logger for product lifecycle events.
// It writes to app.log.
//
//	logger.ProductEvent(ctx, logger.EventProductCreated,
//	    zap.String("product_id", p.ID),
//	    zap.String("seller_id", sellerID)).Info("product created")
func ProductEvent(ctx *gin.Context, event string, extra ...zap.Field) *zap.Logger {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("trace_id", TraceIDFrom(ctx)),
		zap.String("log_type", "product"),
	}
	if userID, ok := ctx.Get("userId"); ok {
		fields = append(fields, zap.Any("actor_id", userID))
	}
	fields = append(fields, extra...)
	return L().With(fields...)
}

// EnquiryEvent returns a logger for enquiry processing events.
//
//	logger.EnquiryEvent(ctx, logger.EventEnquiryAccepted,
//	    zap.String("enquiry_id", id)).Info("enquiry accepted")
func EnquiryEvent(ctx *gin.Context, event string, extra ...zap.Field) *zap.Logger {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("trace_id", TraceIDFrom(ctx)),
		zap.String("log_type", "enquiry"),
	}
	if userID, ok := ctx.Get("userId"); ok {
		fields = append(fields, zap.Any("actor_id", userID))
	}
	fields = append(fields, extra...)
	return L().With(fields...)
}

// PaymentEvent returns the payment sub-logger.  All payment logs additionally
// go to payment.log (see logger.go Init) for PCI-DSS audit trails.
//
//	logger.PaymentEvent(ctx, logger.EventPaymentSuccess,
//	    zap.String("order_id", orderID),
//	    zap.String("gateway", "razorpay"),
//	    zap.Int64("amount_paise", amount)).Info("payment captured")
func PaymentEvent(ctx *gin.Context, event string, extra ...zap.Field) *zap.Logger {
	fields := []zap.Field{
		zap.String("event", event),
		zap.String("trace_id", TraceIDFrom(ctx)),
		zap.String("log_type", "payment"),
	}
	if userID, ok := ctx.Get("userId"); ok {
		fields = append(fields, zap.Any("actor_id", userID))
	}
	fields = append(fields, extra...)
	return Payment().With(fields...)
}

// AdminAction logs a privileged admin CUD operation to the audit log.
// Use this whenever an admin creates, updates, or deletes sensitive data.
//
//	logger.AdminAction(ctx, "department.delete",
//	    zap.String("department_id", id)).Info("admin deleted department")
func AdminAction(ctx *gin.Context, action string, extra ...zap.Field) *zap.Logger {
	fields := []zap.Field{
		zap.String("event", EventSecurityAdminAction),
		zap.String("action", action),
		zap.String("trace_id", TraceIDFrom(ctx)),
		zap.String("client_ip", ctx.ClientIP()),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	}
	if userID, ok := ctx.Get("userId"); ok {
		fields = append(fields, zap.Any("admin_id", userID))
	}
	fields = append(fields, extra...)
	return Audit().With(fields...)
}
