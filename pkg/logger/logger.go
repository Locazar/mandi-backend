// Package logger provides a structured, leveled, multi-output logging framework
// for the mandi-backend service. All output is JSON so it can be ingested by
// Elasticsearch, Loki, Cloud Logging, or any log-aggregation pipeline.
//
// # Log outputs
//
//	stdout        – human-readable (development) or JSON (production)
//	logs/app.log  – JSON, rotated daily, 30-day retention, 5 backups
//
// # Distributed tracing
//
// Every in-process log entry automatically carries the trace_id set by the
// request middleware (see middleware.go). Downstream services should propagate
// the same ID via the X-Trace-ID HTTP header.
package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ── public API ────────────────────────────────────────────────────────────────

var global *zap.Logger

// L returns the global logger. It is safe to call before Init — in that case
// it returns a no-op logger so callers never panic.
func L() *zap.Logger {
	if global == nil {
		return zap.NewNop()
	}
	return global
}

// S returns the global sugared logger for printf-style convenience.
func S() *zap.SugaredLogger { return L().Sugar() }

// Sync flushes any buffered log entries. Call it in a defer in main().
func Sync() { _ = L().Sync() }

// ── initialisation ────────────────────────────────────────────────────────────

// Config controls logger behaviour.
type Config struct {
	// Level is the minimum log level: debug | info | warn | error.
	Level string
	// Development enables coloured, human-readable console output.
	Development bool
	// LogDir is the directory where rotating log files are written.
	// Defaults to "logs".
	LogDir string
	// ServiceName is embedded in every log line as "service".
	ServiceName string
}

// Init initialises the global logger. It must be called once at startup,
// before any route is registered.
func Init(cfg Config) {
	if cfg.LogDir == "" {
		cfg.LogDir = "logs"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "mandi-backend"
	}

	_ = os.MkdirAll(cfg.LogDir, 0o755)

	level := parseLevel(cfg.Level)

	// ── encoders ──────────────────────────────────────────────────────────────
	jsonEnc := zapcore.NewJSONEncoder(productionEncoderConfig())

	var consoleEnc zapcore.Encoder
	if cfg.Development {
		consoleEnc = zapcore.NewConsoleEncoder(developmentEncoderConfig())
	} else {
		consoleEnc = zapcore.NewJSONEncoder(productionEncoderConfig())
	}

	// ── writers ───────────────────────────────────────────────────────────────
	rotatingFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, "app.log"),
		MaxSize:    100,  // MB per file before rotation
		MaxBackups: 5,    // keep last 5 compressed archives
		MaxAge:     30,   // days
		Compress:   true, // gzip archived files
	})

	// Separate file for security / audit events.
	auditFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, "audit.log"),
		MaxSize:    50,
		MaxBackups: 10, // longer retention for audit
		MaxAge:     90,
		Compress:   true,
	})

	// Separate file for payment transactions (PCI-DSS friendly).
	paymentFile := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, "payment.log"),
		MaxSize:    50,
		MaxBackups: 10,
		MaxAge:     90,
		Compress:   true,
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	// ── cores ─────────────────────────────────────────────────────────────────
	// Main core: everything goes to app.log and stdout.
	mainCore := zapcore.NewTee(
		zapcore.NewCore(jsonEnc, rotatingFile, level),
		zapcore.NewCore(consoleEnc, consoleWriter, level),
	)

	// Audit core: security events go to audit.log in addition to main.
	auditCore := zapcore.NewCore(jsonEnc, auditFile, zap.DebugLevel)

	// Payment core: payment events go to payment.log.
	paymentCore := zapcore.NewCore(jsonEnc, paymentFile, zap.DebugLevel)

	// ── named child loggers ───────────────────────────────────────────────────
	// We store them on the package-level so audit.go can reach them.
	global = zap.New(mainCore, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)).
		With(zap.String("service", cfg.ServiceName))

	auditLogger = global.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, auditCore)
	})).With(zap.String("log_type", "audit"))

	paymentLogger = global.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, paymentCore)
	})).With(zap.String("log_type", "payment"))
}

// ── named sub-loggers (used by audit.go) ─────────────────────────────────────

var auditLogger *zap.Logger
var paymentLogger *zap.Logger

func Audit() *zap.Logger {
	if auditLogger == nil {
		return zap.NewNop()
	}
	return auditLogger
}

func Payment() *zap.Logger {
	if paymentLogger == nil {
		return zap.NewNop()
	}
	return paymentLogger
}

// ── encoder configs ───────────────────────────────────────────────────────────

func productionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339Nano),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func developmentEncoderConfig() zapcore.EncoderConfig {
	cfg := productionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	return cfg
}

func parseLevel(s string) zapcore.Level {
	var l zapcore.Level
	if err := l.Set(s); err != nil {
		return zap.InfoLevel
	}
	return l
}
