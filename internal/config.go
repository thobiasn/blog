package blog

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	BaseURL     string
	ContentDir  string
	DBPath      string
	AdminAPIKey string
	BlogURL     string
}

func LoadConfig() Config {
	return Config{
		Port:        envOr("PORT", "8080"),
		BaseURL:     envOr("BASE_URL", "http://localhost:8080"),
		ContentDir:  envOr("CONTENT_DIR", "content"),
		DBPath:      envOr("DB_PATH", "blog.db"),
		AdminAPIKey: os.Getenv("ADMIN_API_KEY"),
		BlogURL:     os.Getenv("BLOG_URL"),
	}
}

func (c Config) isLocal() bool {
	return strings.Contains(c.BaseURL, "localhost")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
