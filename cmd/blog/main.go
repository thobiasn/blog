package main

import (
	"fmt"
	"os"

	"github.com/thobiasn/blog/internal"
)

func main() {
	if len(os.Args) < 2 {
		blog.Dashboard()
		return
	}

	switch os.Args[1] {
	case "serve":
		blog.Serve()
	case "new":
		blog.New(os.Args[2:])
	case "comments":
		blog.Comments(os.Args[2:])
	case "subscribers":
		blog.Subscribers(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: blog [serve|new|comments|subscribers]\n", os.Args[1])
		os.Exit(1)
	}
}
