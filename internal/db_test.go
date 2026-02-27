package blog

import "testing"

func TestOpenDB(t *testing.T) {
	db := testDB(t)

	// Tables should exist — try inserting a comment
	_, err := db.Exec(
		`INSERT INTO comments (post_slug, author, body) VALUES (?, ?, ?)`,
		"test", "alice", "hello",
	)
	if err != nil {
		t.Fatalf("insert into comments: %v", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM comments`).Scan(&count)
	if err != nil {
		t.Fatalf("counting comments: %v", err)
	}
	if count != 1 {
		t.Errorf("got %d comments, want 1", count)
	}
}

func TestOpenDBIdempotent(t *testing.T) {
	db := testDB(t)

	// Calling createTables again should not error
	if err := createTables(db); err != nil {
		t.Fatalf("second createTables() failed: %v", err)
	}
}
