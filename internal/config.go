package blog

import "os"

type Config struct {
	Port       string
	BaseURL    string
	ContentDir string
}

func LoadConfig() Config {
	return Config{
		Port:       envOr("PORT", "8080"),
		BaseURL:    envOr("BASE_URL", "http://localhost:8080"),
		ContentDir: envOr("CONTENT_DIR", "content"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
