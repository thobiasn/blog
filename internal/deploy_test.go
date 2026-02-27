package blog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeployWebhookNotConfigured(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("POST", "/deploy", nil)
	w := httptest.NewRecorder()
	app.handleDeploy(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestDeployWebhookInvalidSignature(t *testing.T) {
	app := testApp(t)
	app.cfg.DeployWebhookSecret = "test-secret"

	req := httptest.NewRequest("POST", "/deploy", strings.NewReader("{}"))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	w := httptest.NewRecorder()
	app.handleDeploy(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestVerifySignature(t *testing.T) {
	secret := "mysecret"
	body := []byte(`{"ref":"refs/heads/main"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !verifySignature(secret, body, sig) {
		t.Error("valid signature should pass")
	}
	if verifySignature(secret, body, "sha256=wrong") {
		t.Error("wrong signature should fail")
	}
	if verifySignature(secret, body, "noshaprefix") {
		t.Error("missing prefix should fail")
	}
}
