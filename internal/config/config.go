package config

import (
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/spf13/viper"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type ServerConfig struct {
    Port int
}

type DBConfig struct {
    DSN string
}

type RedisConfig struct {
    Addr     string
    Password string
    DB       int
}

type JWTConfig struct {
    AccessSecret  string
    RefreshSecret string
    AccessTTL     time.Duration
    RefreshTTL    time.Duration
}

type SMSConfig struct {
    MSG91Key   string
    TwilioSID  string
    TwilioAuth string
}

type Config struct {
    Server ServerConfig
    DB     DBConfig
    Redis  RedisConfig
    JWT    JWTConfig
    SMS    SMSConfig
}

// Load reads environment variables using viper and returns Config
func Load() (*Config, error) {
    viper.SetDefault("SERVER_PORT", 8080)
    viper.SetDefault("ACCESS_TTL_MIN", 15)
    viper.SetDefault("REFRESH_TTL_HOURS", 24*7)

    viper.AutomaticEnv()

    cfg := &Config{
        Server: ServerConfig{Port: viper.GetInt("SERVER_PORT")},
        DB:     DBConfig{DSN: viper.GetString("DATABASE_DSN")},
        Redis: RedisConfig{
            Addr:     viper.GetString("REDIS_ADDR"),
            Password: viper.GetString("REDIS_PASSWORD"),
            DB:       viper.GetInt("REDIS_DB"),
        },
        JWT: JWTConfig{
            AccessSecret:  viper.GetString("JWT_ACCESS_SECRET"),
            RefreshSecret: viper.GetString("JWT_REFRESH_SECRET"),
            AccessTTL:     time.Minute * time.Duration(viper.GetInt("ACCESS_TTL_MIN")),
            RefreshTTL:    time.Hour * time.Duration(viper.GetInt("REFRESH_TTL_HOURS")),
        },
        SMS: SMSConfig{
            MSG91Key:   viper.GetString("MSG91_KEY"),
            TwilioSID:  viper.GetString("TWILIO_SID"),
            TwilioAuth: viper.GetString("TWILIO_AUTH"),
        },
    }

    if cfg.DB.DSN == "" {
        return nil, fmt.Errorf("DATABASE_DSN is required")
    }
    if cfg.Redis.Addr == "" {
        cfg.Redis.Addr = "localhost:6379"
    }
    return cfg, nil
}

func (c *Config) NewGorm() (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(c.DB.DSN), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}

func (c *Config) NewRedis() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     c.Redis.Addr,
        Password: c.Redis.Password,
        DB:       c.Redis.DB,
    })
}
