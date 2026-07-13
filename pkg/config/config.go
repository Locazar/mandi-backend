package config

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
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
	DBSSLRootCert string `mapstructure:"DB_SSL_ROOT_CERT"` // path to CA cert; enables sslmode=verify-full when set

	AdminAuthKey string `mapstructure:"ADMIN_AUTH_KEY"`
	UserAuthKey  string `mapstructure:"USER_AUTH_KEY"`

	TwilioAuthToken  string `mapstructure:"AUTH_TOKEN"`
	TwilioAccountSID string `mapstructure:"ACCOUNT_SID"`
	TwilioServiceID  string `mapstructure:"SERVICE_SID"`

	TwoFactorAPIKey string `mapstructure:"TWO_FACTOR_API_KEY"`

	RazorPayKey           string `mapstructure:"RAZOR_PAY_KEY"`
	RazorPaySecret        string `mapstructure:"RAZOR_PAY_SECRET"`
	RazorPayWebhookSecret string `mapstructure:"RAZORPAY_WEBHOOK_SECRET"`

	// GSTPercentBasisPoints is the current GST rate applied to subscription
	// prices, in basis points (1800 = 18.00%). Prices are GST-inclusive; this
	// rate is snapshotted onto each order at creation. Defaults to 1800.
	GSTPercentBasisPoints int `mapstructure:"GST_PERCENT_BASIS_POINTS"`

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

	// Set to "true" to skip OTP verification (development/testing only)
	SkipOTPValidation bool `mapstructure:"SKIP_OTP_VALIDATION"`

	// SubscriptionGateEnabled toggles customer-facing product-visibility gating:
	// when true, products of sellers without an active subscription are hidden
	// from customer discovery. Shops always stay visible (marked is_subscribed).
	// Default false so the feature ships dark until explicitly enabled.
	SubscriptionGateEnabled bool `mapstructure:"SUBSCRIPTION_GATE_ENABLED"`

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
	"DB_HOST", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_PORT", "DB_SSL_ROOT_CERT", // database
	"ADMIN_AUTH_KEY", "USER_AUTH_KEY", // token auth
	"AUTH_TOKEN", "ACCOUNT_SID", "SERVICE_SID", // twilio
	"TWO_FACTOR_API_KEY",                                           // 2factor.in sms otp
	"RAZOR_PAY_KEY", "RAZOR_PAY_SECRET", "RAZORPAY_WEBHOOK_SECRET", // razor pay
	"GST_PERCENT_BASIS_POINTS",                              // gst rate applied to subscriptions (basis points)
	"STRIPE_SECRET", "STRIPE_PUBLISH_KEY", "STRIPE_WEBHOOK", // stripe
	"GOAUTH_CLIENT_ID", "GOAUTH_CLIENT_SECRET", "GOAUTH_CALL_BACK_URL", //goath
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
	"SKIP_OTP_VALIDATION",
	"SUBSCRIPTION_GATE_ENABLED",
}

func LoadConfig() (config Config, err error) {

	// read from .env file
	viper.AddConfigPath("./")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.SetDefault("GST_PERCENT_BASIS_POINTS", 1800)
	viper.SetDefault("security.http_port", ":3000")
	viper.SetDefault("security.https_port", ":3443")
	viper.SetDefault("security.tls_min_version", "1.2")
	viper.SetDefault("security.tls_max_version", "1.3")
	viper.SetDefault("security.cipher_suites", []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"})
	viper.SetDefault("security.hsts_max_age", 63072000)
	viper.SetDefault("security.cookie_http_only", true)
	viper.SetDefault("security.cookie_same_site", "Lax")
	viper.SetDefault("security.secure_cookies", true)
	viper.SetDefault("security.rate_limit_window_seconds", 60)
	viper.SetDefault("security.rate_limit_requests", 200)
	viper.SetDefault("security.brute_force_max_attempts", 10)
	viper.SetDefault("security.brute_force_window_seconds", 300)
	viper.SetDefault("security.brute_force_block_duration_seconds", 900)
	_ = viper.ReadInConfig()
	// bind from system envs so values can override config file settings
	for _, env := range envsNames {
		if err := viper.BindEnv(env); err != nil {
			return config, err
		}
	}

	_ = viper.BindEnv("security.enable_tls", "SECURITY_ENABLE_TLS")
	_ = viper.BindEnv("security.https_redirect", "SECURITY_HTTPS_REDIRECT")
	_ = viper.BindEnv("security.http_port", "SECURITY_HTTP_PORT")
	_ = viper.BindEnv("security.https_port", "SECURITY_HTTPS_PORT")
	_ = viper.BindEnv("security.tls_cert_file", "SECURITY_TLS_CERT_FILE")
	_ = viper.BindEnv("security.tls_key_file", "SECURITY_TLS_KEY_FILE")
	_ = viper.BindEnv("security.tls_min_version", "SECURITY_TLS_MIN_VERSION")
	_ = viper.BindEnv("security.tls_max_version", "SECURITY_TLS_MAX_VERSION")
	_ = viper.BindEnv("security.cipher_suites", "SECURITY_CIPHER_SUITES")
	_ = viper.BindEnv("security.hsts_max_age", "SECURITY_HSTS_MAX_AGE")
	_ = viper.BindEnv("security.hsts_include_subdomains", "SECURITY_HSTS_INCLUDE_SUBDOMAINS")
	_ = viper.BindEnv("security.hsts_preload", "SECURITY_HSTS_PRELOAD")
	_ = viper.BindEnv("security.secure_cookies", "SECURITY_SECURE_COOKIES")
	_ = viper.BindEnv("security.cookie_http_only", "SECURITY_COOKIE_HTTP_ONLY")
	_ = viper.BindEnv("security.cookie_same_site", "SECURITY_COOKIE_SAME_SITE")
	_ = viper.BindEnv("security.rate_limit_requests", "SECURITY_RATE_LIMIT_REQUESTS")
	_ = viper.BindEnv("security.rate_limit_window_seconds", "SECURITY_RATE_LIMIT_WINDOW_SECONDS")
	_ = viper.BindEnv("security.rate_limiting_enabled", "SECURITY_RATE_LIMITING_ENABLED")
	_ = viper.BindEnv("security.brute_force_protection_enabled", "SECURITY_BRUTE_FORCE_PROTECTION_ENABLED")
	_ = viper.BindEnv("security.brute_force_max_attempts", "SECURITY_BRUTE_FORCE_MAX_ATTEMPTS")
	_ = viper.BindEnv("security.brute_force_window_seconds", "SECURITY_BRUTE_FORCE_WINDOW_SECONDS")
	_ = viper.BindEnv("security.brute_force_block_duration_seconds", "SECURITY_BRUTE_FORCE_BLOCK_DURATION_SECONDS")
	_ = viper.BindEnv("security.config_file", "SECURITY_CONFIG_FILE")

	if jsonPath := viper.GetString("SECURITY_CONFIG_FILE"); jsonPath != "" {
		if fileBytes, fileErr := os.ReadFile(jsonPath); fileErr == nil {
			jsonViper := viper.New()
			jsonViper.SetConfigType("json")
			if err := jsonViper.ReadConfig(bytes.NewBuffer(fileBytes)); err == nil {
				_ = viper.MergeConfigMap(jsonViper.AllSettings())
			}
		}
	} else if _, statErr := os.Stat("config.json"); statErr == nil {
		jsonViper := viper.New()
		jsonViper.SetConfigFile("config.json")
		if err := jsonViper.ReadInConfig(); err == nil {
			_ = viper.MergeConfigMap(jsonViper.AllSettings())
		}
	}

	// support comma-separated cipher suite configuration via env var
	if cipherCSV := viper.GetString("SECURITY_CIPHER_SUITES"); cipherCSV != "" {
		if !strings.Contains(cipherCSV, "[") {
			viper.Set("SECURITY_CIPHER_SUITES", strings.Split(cipherCSV, ","))
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}

	if err := validator.New().Struct(config); err != nil {
		return config, err
	}

	// Ensure Firebase JSON and project id are available in struct config for services
	if config.FirebaseConfig == "" {
		if firebaseCfg := viper.GetString("FIREBASE_CONFIG"); firebaseCfg != "" {
			config.FirebaseConfig = firebaseCfg
		} else if b, marshalErr := json.Marshal(firbaseConfig); marshalErr == nil {
			config.FirebaseConfig = string(b)
		}
	}

	if config.FirebaseProjectID == "" {
		config.FirebaseProjectID = viper.GetString("FIREBASE_PROJECT_ID")
	}

	// Propagate Firebase credentials file path into the OS environment so the
	// Firebase Admin Go SDK can find them via os.Getenv (Viper doesn't do this
	// automatically).
	if credPath := viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credPath)
	}

	// The Firestore watcher mode is read via os.Getenv at startup time.
	// Mirror the .env / Viper value into the process environment so the server
	// can correctly disable its enquiry watcher when Cloud Functions own that flow.
	if mode := strings.Trim(strings.TrimSpace(viper.GetString("ENQUIRY_NOTIFICATION_HANDLER")), `"'`); mode != "" {
		_ = os.Setenv("ENQUIRY_NOTIFICATION_HANDLER", mode)
	}

	if (config.S3Endpoint == "" || config.S3Bucket == "" || config.S3AccessKey == "" || config.S3SecretKey == "") &&
		(config.AwsBucketName != "" || config.AwsAccessKeyID != "") {
		log.Println("[deprecation] AWS_* env vars are being used as fallback for object storage; set S3_* vars before Phase G")
	}

	return config, nil
}
