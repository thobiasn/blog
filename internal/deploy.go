package blog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func (app *App) handleDeploy(w http.ResponseWriter, r *http.Request) {
	secret := app.cfg.DeployWebhookSecret
	if secret == "" {
		http.Error(w, "webhook not configured", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(secret, body, sig) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	go func() {
		app.deployMu.Lock()
		defer app.deployMu.Unlock()

		out, err := exec.Command("git", "pull", "--ff-only").CombinedOutput()
		if err != nil {
			log.Printf("git pull failed: %v\n%s", err, out)
			return
		}
		log.Printf("git pull: %s", out)

		if err := app.reload(); err != nil {
			log.Printf("reload after deploy: %v", err)
		} else {
			log.Println("content reloaded after deploy")
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func verifySignature(secret string, body []byte, sig string) bool {
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig[7:]))
}
