package blog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSubscribeFlow(t *testing.T) {
	app := testApp(t)

	// GET /subscribe shows form
	req := httptest.NewRequest("GET", "/subscribe", nil)
	w := httptest.NewRecorder()
	app.handleSubscribeForm(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /subscribe: status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Subscribe") {
		t.Error("form page should contain Subscribe")
	}

	// POST /subscribe with valid email
	form := url.Values{"email": {"test@example.com"}}
	req = httptest.NewRequest("POST", "/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "127.0.0.1:1234"
	w = httptest.NewRecorder()
	app.handleSubscribe(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /subscribe: status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), "Check your email") {
		t.Error("should show check email message")
	}

	// Verify the subscriber was inserted
	var verifyToken, unsubToken string
	err := app.db.QueryRow(`SELECT verify_token, unsubscribe_token FROM subscribers WHERE email = ?`,
		"test@example.com").Scan(&verifyToken, &unsubToken)
	if err != nil {
		t.Fatalf("querying subscriber: %v", err)
	}
	if verifyToken == "" || unsubToken == "" {
		t.Fatal("tokens should not be empty")
	}

	// Verify subscription
	req = httptest.NewRequest("GET", "/subscribe/verify?token="+verifyToken, nil)
	w = httptest.NewRecorder()
	app.handleSubscribeVerify(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("verify: status = %d, want %d", w.Code, http.StatusOK)
	}

	var verified bool
	app.db.QueryRow(`SELECT verified FROM subscribers WHERE email = ?`, "test@example.com").Scan(&verified)
	if !verified {
		t.Error("subscriber should be verified")
	}

	// Unsubscribe
	req = httptest.NewRequest("GET", "/subscribe/remove?token="+unsubToken, nil)
	w = httptest.NewRecorder()
	app.handleSubscribeRemove(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("remove: status = %d, want %d", w.Code, http.StatusOK)
	}

	var count int
	app.db.QueryRow(`SELECT COUNT(*) FROM subscribers WHERE email = ?`, "test@example.com").Scan(&count)
	if count != 0 {
		t.Error("subscriber should be deleted")
	}
}

func TestSubscribeDuplicateEmail(t *testing.T) {
	app := testApp(t)

	submit := func() int {
		form := url.Values{"email": {"dup@example.com"}}
		req := httptest.NewRequest("POST", "/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		app.handleSubscribe(w, req)
		return w.Code
	}

	if code := submit(); code != http.StatusOK {
		t.Fatalf("first submit: status = %d", code)
	}
	if code := submit(); code != http.StatusOK {
		t.Fatalf("duplicate submit: status = %d, want 200 (no leak)", code)
	}

	var count int
	app.db.QueryRow(`SELECT COUNT(*) FROM subscribers WHERE email = ?`, "dup@example.com").Scan(&count)
	if count != 1 {
		t.Errorf("should have exactly 1 subscriber row, got %d", count)
	}
}

func TestSubscribeInvalidEmail(t *testing.T) {
	app := testApp(t)

	form := url.Values{"email": {"not-an-email"}}
	req := httptest.NewRequest("POST", "/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	app.handleSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/subscribe/verify?token=bogus", nil)
	w := httptest.NewRecorder()
	app.handleSubscribeVerify(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAdminSubscribers(t *testing.T) {
	app := testApp(t)

	// Insert a verified and an unverified subscriber
	app.db.Exec(`INSERT INTO subscribers (email, verified, verify_token, unsubscribe_token) VALUES (?, 1, '', 'u1')`,
		"verified@example.com")
	app.db.Exec(`INSERT INTO subscribers (email, verified, verify_token, unsubscribe_token) VALUES (?, 0, 'v2', 'u2')`,
		"unverified@example.com")

	req := httptest.NewRequest("GET", "/api/admin/subscribers", nil)
	w := httptest.NewRecorder()
	app.handleAdminSubscribers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var result struct {
		Total    int `json:"total"`
		Verified int `json:"verified"`
	}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("total = %d, want 2", result.Total)
	}
	if result.Verified != 1 {
		t.Errorf("verified = %d, want 1", result.Verified)
	}
}

func TestSeedNotifiedPosts(t *testing.T) {
	app := testApp(t)

	app.seedNotifiedPosts()

	var count int
	app.db.QueryRow(`SELECT COUNT(*) FROM notified_posts`).Scan(&count)
	// Only "first-post" is public (draft and private are excluded)
	if count != 1 {
		t.Errorf("notified_posts count = %d, want 1", count)
	}

	// Calling again should be idempotent
	app.seedNotifiedPosts()
	app.db.QueryRow(`SELECT COUNT(*) FROM notified_posts`).Scan(&count)
	if count != 1 {
		t.Errorf("after second seed: count = %d, want 1", count)
	}
}

func TestNotifyNewPostsSkipsAlreadyNotified(t *testing.T) {
	app := testApp(t)

	// Seed existing posts
	app.seedNotifiedPosts()

	// notifyNewPosts with same posts should not crash (SMTP not configured, so no emails sent)
	app.notifyNewPosts(publicPosts(app.posts))

	var count int
	app.db.QueryRow(`SELECT COUNT(*) FROM notified_posts`).Scan(&count)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
