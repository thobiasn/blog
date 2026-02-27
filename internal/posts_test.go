package blog

import (
	"testing"
	"time"
)

func TestPublicPosts(t *testing.T) {
	posts := []Post{
		{Slug: "public-1", Status: "public", Private: false},
		{Slug: "draft-1", Status: "draft", Private: false},
		{Slug: "private-1", Status: "public", Private: true},
		{Slug: "public-2", Status: "public", Private: false},
	}

	got := publicPosts(posts)
	if len(got) != 2 {
		t.Fatalf("publicPosts() returned %d posts, want 2", len(got))
	}
	if got[0].Slug != "public-1" || got[1].Slug != "public-2" {
		t.Errorf("publicPosts() = %v, want public-1 and public-2", got)
	}
}

func TestPublicPostsEmpty(t *testing.T) {
	got := publicPosts(nil)
	if got != nil {
		t.Errorf("publicPosts(nil) = %v, want nil", got)
	}
}

func TestFilterByTag(t *testing.T) {
	posts := []Post{
		{Slug: "a", Tags: []string{"go", "web"}},
		{Slug: "b", Tags: []string{"rust"}},
		{Slug: "c", Tags: []string{"go", "cli"}},
		{Slug: "d", Tags: nil},
	}

	got := filterByTag(posts, "go")
	if len(got) != 2 {
		t.Fatalf("filterByTag(go) returned %d posts, want 2", len(got))
	}
	if got[0].Slug != "a" || got[1].Slug != "c" {
		t.Errorf("filterByTag(go) = %v, want a and c", got)
	}

	got = filterByTag(posts, "nonexistent")
	if got != nil {
		t.Errorf("filterByTag(nonexistent) = %v, want nil", got)
	}
}

func TestFindPost(t *testing.T) {
	posts := []Post{
		{Slug: "first", Title: "First"},
		{Slug: "second", Title: "Second"},
	}

	p, ok := findPost(posts, "second")
	if !ok {
		t.Fatal("findPost(second) not found")
	}
	if p.Title != "Second" {
		t.Errorf("findPost(second).Title = %q, want %q", p.Title, "Second")
	}

	_, ok = findPost(posts, "nope")
	if ok {
		t.Error("findPost(nope) should not be found")
	}

	_, ok = findPost(nil, "any")
	if ok {
		t.Error("findPost on nil should not be found")
	}
}

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter()

	// Should allow up to limit (5)
	for i := 0; i < 5; i++ {
		if !rl.allow("test-ip") {
			t.Fatalf("allow() returned false on request %d, want true", i+1)
		}
	}

	// 6th should be denied
	if rl.allow("test-ip") {
		t.Error("allow() returned true on 6th request, want false")
	}

	// Different key should still be allowed
	if !rl.allow("other-ip") {
		t.Error("allow(other-ip) returned false, want true")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    2,
		window:   time.Millisecond,
	}

	rl.allow("ip")
	rl.allow("ip")
	if rl.allow("ip") {
		t.Error("allow() should be denied at limit")
	}

	// Wait for window to expire
	time.Sleep(2 * time.Millisecond)

	if !rl.allow("ip") {
		t.Error("allow() should succeed after window expiry")
	}
}
