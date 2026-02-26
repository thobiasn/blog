package main

import (
	"fmt"
	"os"

	"github.com/thobiasn/blog/internal"
)

func main() {
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "serve":
		blog.Serve()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nusage: blog serve\n", cmd)
		os.Exit(1)
	}
}
