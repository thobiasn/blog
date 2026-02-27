package blog

import (
	"strings"
	"testing"
)

func TestSendMailNotConfigured(t *testing.T) {
	app := testApp(t)

	err := app.sendMail("test@example.com", "Test", "Hello")
	if err == nil {
		t.Fatal("expected error when SMTP not configured")
	}
	if !strings.Contains(err.Error(), "SMTP not configured") {
		t.Errorf("error = %q, want SMTP not configured", err)
	}
}
