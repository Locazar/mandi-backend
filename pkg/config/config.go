package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// to store env variables
type Config struct {
	AdminEmail    string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword string `mapstructure:"ADMIN_PASSWORD"`
	DBHost        string `mapstructure:"DB_HOST"`
	DBName        string `mapstructure:"DB_NAME"`
	DBUser        string `mapstructure:"DB_USER"`
	DBPassword    string `mapstructure:"DB_PASSWORD"`
	DBPort        string `mapstructure:"DB_PORT"`

	AdminAuthKey string `mapstructure:"ADMIN_AUTH_KEY"`
	UserAuthKey  string `mapstructure:"USER_AUTH_KEY"`

	TwilioAuthToken  string `mapstructure:"AUTH_TOKEN"`
	TwilioAccountSID string `mapstructure:"ACCOUNT_SID"`
	TwilioServiceID  string `mapstructure:"SERVICE_SID"`

	RazorPayKey           string `mapstructure:"RAZOR_PAY_KEY"`
	RazorPaySecret        string `mapstructure:"RAZOR_PAY_SECRET"`
	RazorPayWebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`

	StripSecretKey      string `mapstructure:"STRIPE_SECRET"`
	StripPublishKey     string `mapstructure:"STRIPE_PUBLISH_KEY"`
	StripeWebhookSecret string `mapstructure:"STRIPE_WEBHOOK"`

	GoathClientID      string `mapstructure:"GOAUTH_CLIENT_ID"`
	GoauthClientSecret string `mapstructure:"GOAUTH_CLIENT_SECRET"`
	GoauthCallbackUrl  string `mapstructure:"GOAUTH_CALL_BACK_URL"`

	// Legacy AWS_* — kept as fallback for one release; remove after Phase G ships.
	AwsAccessKeyID string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AwsSecretKey   string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion      string `mapstructure:"AWS_REGION"`
	AwsBucketName  string `mapstructure:"AWS_BUCKET_NAME"`

	// Object Storage (Utho S3-compatible). Takes precedence over AWS_* above.
	S3Endpoint      string `mapstructure:"S3_ENDPOINT"`
	S3Region        string `mapstructure:"S3_REGION"`
	S3Bucket        string `mapstructure:"S3_BUCKET"`
	S3AccessKey     string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey     string `mapstructure:"S3_SECRET_KEY"`
	S3PublicBaseURL string `mapstructure:"S3_PUBLIC_BASE_URL"`

	SharedUploadsPath string `mapstructure:"SHARED_UPLOADS_PATH"`

	ElasticsearchURL string `mapstructure:"ELASTICSEARCH_URL"`

	AIServiceURL string `mapstructure:"AI_SERVICE_URL"`

	FirebaseProjectID          string `mapstructure:"FIREBASE_PROJECT_ID"`
	FirebaseConfig             string `mapstructure:"FIREBASE_CONFIG"`
	EnquiryNotificationHandler string `mapstructure:"ENQUIRY_NOTIFICATION_HANDLER"`

	// PII encryption keyring. PIIEncryptionKeys is a comma-separated list of
	// "<id>:<base64-32-byte-key>" entries; PIIEncryptionActiveKey names the key
	// used for new writes. Older keys remain present to decrypt existing data.
	PIIEncryptionKeys      string `mapstructure:"PII_ENCRYPTION_KEYS"`
	PIIEncryptionActiveKey string `mapstructure:"PII_ENCRYPTION_ACTIVE_KEY"`

	Security SecurityConfig `mapstructure:"security"`
}

type SecurityConfig struct {
	EnableTLS                      bool     `mapstructure:"enable_tls"`
	HTTPSRedirect                  bool     `mapstructure:"https_redirect"`
	HTTPPort                       string   `mapstructure:"http_port"`
	HTTPSPort                      string   `mapstructure:"https_port"`
	TLSCertFile                    string   `mapstructure:"tls_cert_file"`
	TLSKeyFile                     string   `mapstructure:"tls_key_file"`
	TLSMinVersion                  string   `mapstructure:"tls_min_version"`
	TLSMaxVersion                  string   `mapstructure:"tls_max_version"`
	CipherSuites                   []string `mapstructure:"cipher_suites"`
	HSTSMaxAge                     int      `mapstructure:"hsts_max_age"`
	HSTSIncludeSubDomains          bool     `mapstructure:"hsts_include_subdomains"`
	HSTSPreload                    bool     `mapstructure:"hsts_preload"`
	SecureCookies                  bool     `mapstructure:"secure_cookies"`
	CookieHTTPOnly                 bool     `mapstructure:"cookie_http_only"`
	CookieSameSite                 string   `mapstructure:"cookie_same_site"`
	RateLimitingEnabled            bool     `mapstructure:"rate_limiting_enabled"`
	RateLimitRequests              int      `mapstructure:"rate_limit_requests"`
	RateLimitWindowSeconds         int      `mapstructure:"rate_limit_window_seconds"`
	BruteForceProtectionEnabled    bool     `mapstructure:"brute_force_protection_enabled"`
	BruteForceMaxAttempts          int      `mapstructure:"brute_force_max_attempts"`
	BruteForceWindowSeconds        int      `mapstructure:"brute_force_window_seconds"`
	BruteForceBlockDurationSeconds int      `mapstructure:"brute_force_block_duration_seconds"`
	SecurityConfigFile             string   `mapstructure:"config_file"`
}

var firbaseConfig = map[string]interface{}{
	"type":                        "service_account",
	"project_id":                  "mandi-backend-379522",
	"private_key_id":              "some_key_id",
	"private_key":                 "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
	"client_email":                "some_email@mandi-backend-379522.iam.gserviceaccount.com",
	"client_id":                   "some_client_id",
	"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
	"token_uri":                   "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url":        "some_cert_url",
}

// name of envs and used to read from system envs
var envsNames = []string{
	"ADMIN_EMAIL", "ADMIN_PASSWORD",
	"DB_HOST", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_PORT", // database
	"ADMIN_AUTH_KEY", "USER_AUTH_KEY", // token auth
	"AUTH_TOKEN", "ACCOUNT_SID", "SERVICE_SID", // twilio
	"RAZOR_PAY_KEY", "RAZOR_PAY_SECRET", "RAZORPAY_WEBHOOK_SECRET", // razor pay
	"STRIPE_SECRET", "STRIPE_PUBLISH_KEY", "STRIPE_WEBHOOK", // stripe
	"GOAUTH_CLIENT_ID", "GOAUTH_CLIENT_SECRET", "GOAUTH_CALL_BACK_URL", // goath
	"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "AWS_BUCKET_NAME", // legacy aws fallback
	"S3_ENDPOINT", "S3_REGION", "S3_BUCKET", "S3_ACCESS_KEY", "S3_SECRET_KEY", "S3_PUBLIC_BASE_URL", // object storage
	"SHARED_UPLOADS_PATH", // shared uploads directory (deprecated)
	"ELASTICSEARCH_URL",   // elasticsearch
	"AI_SERVICE_URL",      // ai service
	// Firebase — either an ADC credentials file path or inline JSON
	"GOOGLE_APPLICATION_CREDENTIALS",
	"FIREBASE_CONFIG",
	"FIREBASE_PROJECT_ID",
	"ENQUIRY_NOTIFICATION_HANDLER",
	"SECURITY_CONFIG_FILE",
	"SECURITY_ENABLE_TLS",
	"SECURITY_HTTPS_REDIRECT",
	"SECURITY_HTTP_PORT",
	"SECURITY_HTTPS_PORT",
	"SECURITY_TLS_CERT_FILE",
	"SECURITY_TLS_KEY_FILE",
	"SECURITY_TLS_MIN_VERSION",
	"SECURITY_TLS_MAX_VERSION",
	"SECURITY_CIPHER_SUITES",
	"SECURITY_HSTS_MAX_AGE",
	"SECURITY_HSTS_INCLUDE_SUBDOMAINS",
	"SECURITY_HSTS_PRELOAD",
	"SECURITY_SECURE_COOKIES",
	"SECURITY_COOKIE_HTTP_ONLY",
	"SECURITY_COOKIE_SAME_SITE",
	"SECURITY_RATE_LIMITING_ENABLED",
	"SECURITY_RATE_LIMIT_REQUESTS",
	"SECURITY_RATE_LIMIT_WINDOW_SECONDS",
	"SECURITY_BRUTE_FORCE_PROTECTION_ENABLED",
	"SECURITY_BRUTE_FORCE_MAX_ATTEMPTS",
	"SECURITY_BRUTE_FORCE_WINDOW_SECONDS",
	"SECURITY_BRUTE_FORCE_BLOCK_DURATION_SECONDS",
	"PII_ENCRYPTION_KEYS",
	"PII_ENCRYPTION_ACTIVE_KEY",
}

func LoadConfig() (config Config, err error) {
	// Load .env file into OS environment (auto-load, silently skips if missing)
	_ = godotenv.Load(".env")

	// Directly read from OS environment variables since godotenv loads them there
	config.AdminEmail = os.Getenv("ADMIN_EMAIL")
	config.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	config.DBHost = os.Getenv("DB_HOST")
	config.DBName = os.Getenv("DB_NAME")
	config.DBUser = os.Getenv("DB_USER")
	config.DBPassword = os.Getenv("DB_PASSWORD")
	config.DBPort = os.Getenv("DB_PORT")
	config.AdminAuthKey = os.Getenv("ADMIN_AUTH_KEY")
	config.UserAuthKey = os.Getenv("USER_AUTH_KEY")
	config.TwilioAuthToken = os.Getenv("AUTH_TOKEN")
	config.TwilioAccountSID = os.Getenv("ACCOUNT_SID")
	config.TwilioServiceID = os.Getenv("SERVICE_SID")
	config.RazorPayKey = os.Getenv("RAZOR_PAY_KEY")
	config.RazorPaySecret = os.Getenv("RAZOR_PAY_SECRET")
	config.RazorPayWebhookSecret = os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	config.StripSecretKey = os.Getenv("STRIPE_SECRET")
	config.StripPublishKey = os.Getenv("STRIPE_PUBLISH_KEY")
	config.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK")
	config.GoathClientID = os.Getenv("GOAUTH_CLIENT_ID")
	config.GoauthClientSecret = os.Getenv("GOAUTH_CLIENT_SECRET")
	config.GoauthCallbackUrl = os.Getenv("GOAUTH_CALL_BACK_URL")
	config.AwsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AwsSecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	config.AwsRegion = os.Getenv("AWS_REGION")
	config.AwsBucketName = os.Getenv("AWS_BUCKET_NAME")
	config.S3Endpoint = os.Getenv("S3_ENDPOINT")
	config.S3Region = os.Getenv("S3_REGION")
	config.S3Bucket = os.Getenv("S3_BUCKET")
	config.S3AccessKey = os.Getenv("S3_ACCESS_KEY")
	config.S3SecretKey = os.Getenv("S3_SECRET_KEY")
	config.S3PublicBaseURL = os.Getenv("S3_PUBLIC_BASE_URL")
	config.SharedUploadsPath = os.Getenv("SHARED_UPLOADS_PATH")
	config.ElasticsearchURL = os.Getenv("ELASTICSEARCH_URL")
	config.AIServiceURL = os.Getenv("AI_SERVICE_URL")
	config.FirebaseProjectID = os.Getenv("FIREBASE_PROJECT_ID")
	config.FirebaseConfig = os.Getenv("FIREBASE_CONFIG")
	config.EnquiryNotificationHandler = os.Getenv("ENQUIRY_NOTIFICATION_HANDLER")
	config.PIIEncryptionKeys = os.Getenv("PII_ENCRYPTION_KEYS")
	config.PIIEncryptionActiveKey = os.Getenv("PII_ENCRYPTION_ACTIVE_KEY")

	// Set security defaults and read from environment
	config.Security.HTTPPort = getEnvOrDefault("SECURITY_HTTP_PORT", ":3000")
	config.Security.HTTPSPort = getEnvOrDefault("SECURITY_HTTPS_PORT", ":3443")
	config.Security.TLSMinVersion = getEnvOrDefault("SECURITY_TLS_MIN_VERSION", "1.2")
	config.Security.TLSMaxVersion = getEnvOrDefault("SECURITY_TLS_MAX_VERSION", "1.3")
	config.Security.CipherSuites = []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"}
	config.Security.HSTSMaxAge = 63072000
	config.Security.CookieHTTPOnly = true
	config.Security.CookieSameSite = "Lax"
	config.Security.SecureCookies = true
	config.Security.RateLimitWindowSeconds = 60
	config.Security.RateLimitRequests = 200
	config.Security.BruteForceMaxAttempts = 10
	config.Security.BruteForceWindowSeconds = 300
	config.Security.BruteForceBlockDurationSeconds = 900

	config.Security.EnableTLS = parseEnvBool("SECURITY_ENABLE_TLS", false)
	config.Security.HTTPSRedirect = parseEnvBool("SECURITY_HTTPS_REDIRECT", false)
	config.Security.TLSCertFile = os.Getenv("SECURITY_TLS_CERT_FILE")
	config.Security.TLSKeyFile = os.Getenv("SECURITY_TLS_KEY_FILE")
	config.Security.HSTSIncludeSubDomains = parseEnvBool("SECURITY_HSTS_INCLUDE_SUBDOMAINS", false)
	config.Security.HSTSPreload = parseEnvBool("SECURITY_HSTS_PRELOAD", false)
	config.Security.RateLimitingEnabled = parseEnvBool("SECURITY_RATE_LIMITING_ENABLED", false)
	config.Security.BruteForceProtectionEnabled = parseEnvBool("SECURITY_BRUTE_FORCE_PROTECTION_ENABLED", false)
	config.Security.SecurityConfigFile = os.Getenv("SECURITY_CONFIG_FILE")

	// Set Firebase environment variable for the SDK
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credPath)
	}

	// Mirror the Enquiry Notification Handler into the process environment
	if mode := os.Getenv("ENQUIRY_NOTIFICATION_HANDLER"); mode != "" {
		_ = os.Setenv("ENQUIRY_NOTIFICATION_HANDLER", strings.Trim(strings.TrimSpace(mode), `"'`))
	}

	// Ensure Firebase JSON and project id are available
	if config.FirebaseConfig == "" {
		if b, marshalErr := json.Marshal(firbaseConfig); marshalErr == nil {
			config.FirebaseConfig = string(b)
		}
	}

	if (config.S3Endpoint == "" || config.S3Bucket == "" || config.S3AccessKey == "" || config.S3SecretKey == "") &&
		(config.AwsBucketName != "" || config.AwsAccessKeyID != "") {
		log.Println("[deprecation] AWS_* env vars are being used as fallback for object storage; set S3_* vars before Phase G")
	}

	// ===== PRODUCTION VALIDATION =====
	// Validate all required variables are set (fail fast at startup)
	if err := validateRequiredConfig(config); err != nil {
		return config, err
	}

	return config, nil
}

// validateRequiredConfig checks that critical environment variables are set
// This prevents silent failures at runtime due to missing configuration
func validateRequiredConfig(config Config) error {
	requiredVars := map[string]string{
		"DB_HOST":                   config.DBHost,
		"DB_NAME":                   config.DBName,
		"DB_USER":                   config.DBUser,
		"DB_PASSWORD":               config.DBPassword,
		"DB_PORT":                   config.DBPort,
		"ADMIN_AUTH_KEY":            config.AdminAuthKey,
		"USER_AUTH_KEY":             config.UserAuthKey,
		"PII_ENCRYPTION_KEYS":       config.PIIEncryptionKeys,
		"PII_ENCRYPTION_ACTIVE_KEY": config.PIIEncryptionActiveKey,
	}

	var missingVars []string
	for varName, varValue := range requiredVars {
		if strings.TrimSpace(varValue) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}
