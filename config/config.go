package config

import (
	"os"
)

type Config struct {
	BloggoAPIURL   string
	BloggoAPIToken string
	WebhookSecret  string
	ServerPort     string
	TemplatesDir   string
	StaticDir      string
	PrerenderedDir string
}

func Load() *Config {
	return &Config{
		BloggoAPIURL:   getEnv("BLOGGO_API_URL", "http://localhost:3000/api"),
		BloggoAPIToken: getEnv("BLOGGO_API_TOKEN", ""),
		WebhookSecret:  getEnv("WEBHOOK_SECRET", ""),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		TemplatesDir:   getEnv("TEMPLATES_DIR", "templates"),
		StaticDir:      getEnv("STATIC_DIR", "static"),
		PrerenderedDir: getEnv("PRERENDERED_DIR", "prerendered"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
