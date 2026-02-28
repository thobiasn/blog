package blog

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCommentsBySlug(t *testing.T) {
	app := testApp(t)

	seedComments(t, app.db, []Comment{
		{PostSlug: "first-post", Author: "Alice", Body: "Great post!", Visible: true},
		{PostSlug: "first-post", Author: "Bob", Body: "Hidden", Visible: false},
		{PostSlug: "other-post", Author: "Eve", Body: "Wrong post", Visible: true},
	})

	comments := app.commentsBySlug("first-post")
	if len(comments) != 1 {
		t.Fatalf("got %d comments, want 1 (only visible for this slug)", len(comments))
	}
	if comments[0].Author != "Alice" {
		t.Errorf("Author = %q, want %q", comments[0].Author, "Alice")
	}
}

func TestCommentsBySlugNoDB(t *testing.T) {
	app := testApp(t)
	app.db = nil

	comments := app.commentsBySlug("first-post")
	if comments != nil {
		t.Errorf("got %v, want nil when db is nil", comments)
	}
}

func TestCreateComment(t *testing.T) {
	app := testApp(t)

	err := app.createComment("first-post", "Alice", "Nice!")
	if err != nil {
		t.Fatalf("createComment: %v", err)
	}

	comments := app.commentsBySlug("first-post")
	if len(comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(comments))
	}
	if comments[0].Author != "Alice" || comments[0].Body != "Nice!" {
		t.Errorf("got comment %+v, want Alice/Nice!", comments[0])
	}
}

func TestHandleCommentSubmit(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {"Alice"},
		"body":   {"Great post!"},
	}
	req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "first-post")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); loc != "/posts/first-post#comments" {
		t.Errorf("Location = %q, want /posts/first-post#comments", loc)
	}

	// Comment should be in DB
	comments := app.commentsBySlug("first-post")
	if len(comments) != 1 {
		t.Fatalf("got %d comments, want 1", len(comments))
	}
}

func TestHandleCommentSubmitEmpty(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {""},
		"body":   {""},
	}
	req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "first-post")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCommentSubmitTooLong(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {strings.Repeat("a", 101)},
		"body":   {"valid body"},
	}
	req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "first-post")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCommentSubmitHoneypot(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {"Bot"},
		"body":   {"Buy stuff!"},
		"url":    {"https://spam.example.com"},
	}
	req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "first-post")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	// Should redirect (pretend success) but not save
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	comments := app.commentsBySlug("first-post")
	if len(comments) != 0 {
		t.Error("honeypot triggered but comment was saved")
	}
}

func TestHandleCommentSubmitRateLimit(t *testing.T) {
	app := testApp(t)

	for i := 0; i < 5; i++ {
		form := url.Values{
			"author": {"Spammer"},
			"body":   {"Comment"},
		}
		req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("slug", "first-post")
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()

		app.handleCommentSubmit(w, req)

		if w.Code != http.StatusSeeOther {
			t.Fatalf("request %d: status = %d, want %d", i+1, w.Code, http.StatusSeeOther)
		}
	}

	// 6th should be rate limited
	form := url.Values{
		"author": {"Spammer"},
		"body":   {"One too many"},
	}
	req := httptest.NewRequest("POST", "/posts/first-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "first-post")
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestHandleCommentSubmitPostNotFound(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {"Alice"},
		"body":   {"Comment on nothing"},
	}
	req := httptest.NewRequest("POST", "/posts/nonexistent/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "nonexistent")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleCommentSubmitPrivatePost(t *testing.T) {
	app := testApp(t)

	form := url.Values{
		"author": {"Alice"},
		"body":   {"Comment on private post"},
	}
	req := httptest.NewRequest("POST", "/posts/private-post/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("slug", "private-post")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	app.handleCommentSubmit(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (comments blocked on private posts)", w.Code, http.StatusNotFound)
	}
}
