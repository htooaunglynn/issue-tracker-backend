package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port     string
	Env      string
	LogLevel string
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleTime    time.Duration
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	Driver   string // "smtp", "log", or "noop"
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:     getEnv("PORT", ":8080"),
			Env:      getEnv("ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			URL:            getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/issuetracker"),
			MaxConnections: getEnvInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleTime:    time.Duration(getEnvInt("DB_MAX_IDLE_TIME", 5)) * time.Minute,
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key"),
			AccessExpiry:  time.Duration(getEnvInt("JWT_ACCESS_EXPIRY_HOURS", 1)) * time.Hour,
			RefreshExpiry: time.Duration(getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)) * 24 * time.Hour,
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnv("SMTP_PORT", "1025"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@issuetracker.local"),
			Driver:   getEnv("EMAIL_DRIVER", "log"),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}, nil
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err != nil {
			return defaultVal
		}
		return i
	}
	return defaultVal
}
