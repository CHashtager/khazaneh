package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppAddr                string
	DatabaseURL            string
	TelegramBotToken       string
	AllowedTelegramUserIDs map[int64]bool
	DefaultTimezone        string
	DefaultCurrency        string
	DashboardUsername      string
	DashboardPassword      string
}

func Load() (Config, error) {
	cfg := Config{
		AppAddr:                env("APP_ADDR", ":8080"),
		DatabaseURL:            env("DATABASE_URL", ""),
		TelegramBotToken:       env("TELEGRAM_BOT_TOKEN", ""),
		AllowedTelegramUserIDs: map[int64]bool{},
		DefaultTimezone:        env("DEFAULT_TIMEZONE", "Asia/Tehran"),
		DefaultCurrency:        env("DEFAULT_CURRENCY", "TOMAN"),
		DashboardUsername:      env("DASHBOARD_USERNAME", "admin"),
		DashboardPassword:      env("DASHBOARD_PASSWORD", "change-me"),
	}

	for _, raw := range strings.Split(os.Getenv("ALLOWED_TELEGRAM_USER_IDS"), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return cfg, err
		}
		cfg.AllowedTelegramUserIDs[id] = true
	}

	if cfg.DatabaseURL == "" {
		return cfg, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
