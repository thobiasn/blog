package blog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAdmin(t *testing.T) {
	app := testApp(t)

	called := false
	handler := app.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if !called {
		t.Error("inner handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRequireAdminNoToken(t *testing.T) {
	app := testApp(t)

	called := false
	handler := app.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if called {
		t.Error("inner handler should not be called without token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAdminWrongToken(t *testing.T) {
	app := testApp(t)

	called := false
	handler := app.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if called {
		t.Error("inner handler should not be called with wrong token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAdminNoKey(t *testing.T) {
	app := testApp(t)
	app.cfg.AdminAPIKey = ""

	called := false
	handler := app.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer some-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if called {
		t.Error("inner handler should not be called when admin not configured")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestHandleAdminStats(t *testing.T) {
	app := testApp(t)

	seedComments(t, app.db, []Comment{
		{PostSlug: "first-post", Author: "Alice", Body: "Hi", Visible: true},
		{PostSlug: "first-post", Author: "Bob", Body: "Hey", Visible: true},
	})

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	w := httptest.NewRecorder()

	app.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var stats adminStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if stats.PublicPosts != 1 {
		t.Errorf("PublicPosts = %d, want 1", stats.PublicPosts)
	}
	if stats.PrivatePosts != 1 {
		t.Errorf("PrivatePosts = %d, want 1", stats.PrivatePosts)
	}
	if stats.Comments != 2 {
		t.Errorf("Comments = %d, want 2", stats.Comments)
	}
}

func TestHandleAdminComments(t *testing.T) {
	app := testApp(t)

	seedComments(t, app.db, []Comment{
		{PostSlug: "first-post", Author: "Alice", Body: "Hello", Visible: true},
		{PostSlug: "first-post", Author: "Bob", Body: "World", Visible: false},
	})

	req := httptest.NewRequest("GET", "/api/admin/comments", nil)
	w := httptest.NewRecorder()

	app.handleAdminComments(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var comments []Comment
	if err := json.NewDecoder(w.Body).Decode(&comments); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("got %d comments, want 2 (admin sees all)", len(comments))
	}
}

func TestHandleAdminCommentsEmpty(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("GET", "/api/admin/comments", nil)
	w := httptest.NewRecorder()

	app.handleAdminComments(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Should be [] not null
	body := w.Body.String()
	if body != "[]\n" {
		t.Errorf("body = %q, want %q", body, "[]\n")
	}
}

func TestHandleAdminCommentToggle(t *testing.T) {
	app := testApp(t)

	seedComments(t, app.db, []Comment{
		{PostSlug: "first-post", Author: "Alice", Body: "Hello", Visible: true},
	})

	// Toggle visibility off
	req := httptest.NewRequest("POST", "/api/admin/comments/1/toggle", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	app.handleAdminCommentToggle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Should be hidden now
	var visible bool
	app.db.QueryRow(`SELECT visible FROM comments WHERE id = 1`).Scan(&visible)
	if visible {
		t.Error("comment should be hidden after toggle")
	}

	// Toggle back on
	req = httptest.NewRequest("POST", "/api/admin/comments/1/toggle", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()

	app.handleAdminCommentToggle(w, req)

	app.db.QueryRow(`SELECT visible FROM comments WHERE id = 1`).Scan(&visible)
	if !visible {
		t.Error("comment should be visible after second toggle")
	}
}

func TestHandleAdminCommentToggleNotFound(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("POST", "/api/admin/comments/999/toggle", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	app.handleAdminCommentToggle(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleAdminCommentDelete(t *testing.T) {
	app := testApp(t)

	seedComments(t, app.db, []Comment{
		{PostSlug: "first-post", Author: "Alice", Body: "Hello", Visible: true},
	})

	req := httptest.NewRequest("POST", "/api/admin/comments/1/delete", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	app.handleAdminCommentDelete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var count int
	app.db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&count)
	if count != 0 {
		t.Errorf("got %d comments, want 0 after delete", count)
	}
}

func TestHandleAdminCommentDeleteNotFound(t *testing.T) {
	app := testApp(t)

	req := httptest.NewRequest("POST", "/api/admin/comments/999/delete", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	app.handleAdminCommentDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
