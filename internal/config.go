package blog

import (
	"os"
	"strings"
)

type Config struct {
	Port                string
	BaseURL             string
	ContentDir          string
	DBPath              string
	AdminAPIKey         string
	BlogURL             string
	SMTPHost            string
	SMTPPort            string
	SMTPUsername        string
	SMTPPassword        string
	FromEmail           string
	DeployWebhookSecret string
}

func LoadConfig() Config {
	return Config{
		Port:                envOr("PORT", "8080"),
		BaseURL:             envOr("BASE_URL", "http://localhost:8080"),
		ContentDir:          envOr("CONTENT_DIR", "content"),
		DBPath:              envOr("DB_PATH", "blog.db"),
		AdminAPIKey:         os.Getenv("ADMIN_API_KEY"),
		BlogURL:             os.Getenv("BLOG_URL"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            envOr("SMTP_PORT", "587"),
		SMTPUsername:        os.Getenv("SMTP_USERNAME"),
		SMTPPassword:        os.Getenv("SMTP_PASSWORD"),
		FromEmail:           os.Getenv("FROM_EMAIL"),
		DeployWebhookSecret: os.Getenv("DEPLOY_WEBHOOK_SECRET"),
	}
}

func (c Config) smtpConfigured() bool {
	return c.SMTPHost != "" && c.FromEmail != ""
}

func (c Config) isLocal() bool {
	return strings.HasPrefix(c.BaseURL, "http://localhost")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
