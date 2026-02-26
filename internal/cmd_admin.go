package blog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func adminRequest(method, path string) (*http.Response, error) {
	cfg := LoadConfig()
	if cfg.BlogURL == "" || cfg.AdminAPIKey == "" {
		return nil, fmt.Errorf("BLOG_URL and ADMIN_API_KEY must be set")
	}

	req, err := http.NewRequest(method, cfg.BlogURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AdminAPIKey)
	return http.DefaultClient.Do(req)
}

func Dashboard() {
	resp, err := adminRequest("GET", "/api/admin/stats")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var stats adminStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		fmt.Fprintf(os.Stderr, "error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Blog Dashboard")
	fmt.Println("──────────────")
	fmt.Printf("Public posts:  %d\n", stats.PublicPosts)
	fmt.Printf("Draft posts:   %d\n", stats.DraftPosts)
	fmt.Printf("Private posts: %d\n", stats.PrivatePosts)
	fmt.Printf("Comments:      %d\n", stats.Comments)
}

func Comments(args []string) {
	if len(args) == 0 {
		listComments()
		return
	}

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: blog comments [delete|toggle] <id>")
		os.Exit(1)
	}

	action, id := args[0], args[1]
	switch action {
	case "delete":
		moderateComment("delete", id)
	case "toggle":
		moderateComment("toggle", id)
	default:
		fmt.Fprintf(os.Stderr, "unknown action: %s\nusage: blog comments [delete|toggle] <id>\n", action)
		os.Exit(1)
	}
}

func listComments() {
	resp, err := adminRequest("GET", "/api/admin/comments")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		fmt.Fprintf(os.Stderr, "error decoding response: %v\n", err)
		os.Exit(1)
	}

	if len(comments) == 0 {
		fmt.Println("No comments.")
		return
	}

	for _, c := range comments {
		vis := "visible"
		if !c.Visible {
			vis = "hidden"
		}
		fmt.Printf("#%d [%s] on %s by %s (%s)\n  %s\n\n",
			c.ID, vis, c.PostSlug, c.Author, c.CreatedAt.Format("2006-01-02 15:04"), c.Body)
	}
}

func moderateComment(action, id string) {
	resp, err := adminRequest("POST", "/api/admin/comments/"+id+"/"+action)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "error: %s %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	fmt.Printf("comment %s: %sd\n", id, action)
}

func Subscribers(args []string) {
	fmt.Println("No subscribers yet. (Subscriber management coming in Phase 3)")
}
