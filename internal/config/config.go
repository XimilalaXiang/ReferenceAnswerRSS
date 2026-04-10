package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         int
	DatabasePath string
	JWTSecret    string
	SyncInterval time.Duration

	XinzhiAPIBase string
	XinzhiToken   string
	AuthorID      string

	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	return &Config{
		Port:          getEnvInt("PORT", 8080),
		DatabasePath:  getEnv("DATABASE_PATH", "./data/rss.db"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		SyncInterval:  time.Duration(getEnvInt("SYNC_INTERVAL_HOURS", 12)) * time.Hour,
		XinzhiAPIBase: getEnv("XINZHI_API_BASE", "https://api.xinzhi.zone/api"),
		XinzhiToken:   getEnv("XINZHI_TOKEN", ""),
		AuthorID:      getEnv("AUTHOR_ID", "6905098d5f77b11d2fb2b653"),
		AdminUsername:  getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", "admin"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
