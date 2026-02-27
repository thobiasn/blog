package blog

import "testing"

func TestPostSlug(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"2026-02-25-hello-world.md", "hello-world"},
		{"hello-world.md", "hello-world"},
		{"2026-02-25-a.md", "a"},
		{"no-date-prefix.md", "no-date-prefix"},
		{"short.md", "short"},
		{"2026-12-31-year-end-review.md", "year-end-review"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := postSlug(tt.filename)
			if got != tt.want {
				t.Errorf("postSlug(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
