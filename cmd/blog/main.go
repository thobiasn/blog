package main

import (
	"fmt"
	"os"

	"github.com/thobiasn/blog/internal"
)

const usage = `usage: blog <command>

Commands:
  serve                          start HTTP server
  new post <title>               create a new post (in content/private/)
  new project <name>             create a new project
  publish <slug>                 move post from private to public
  dash                           admin dashboard
  comments                       list recent comments
  comments delete <id>           delete a comment
  comments toggle <id>           toggle comment visibility
  subscribers                    subscriber stats
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return
	}

	switch os.Args[1] {
	case "serve":
		blog.Serve()
	case "new":
		blog.New(os.Args[2:])
	case "publish":
		blog.Publish(os.Args[2:])
	case "dash":
		blog.Dashboard()
	case "comments":
		blog.Comments(os.Args[2:])
	case "subscribers":
		blog.Subscribers(os.Args[2:])
	case "help", "-help", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n%s", os.Args[1], usage)
		os.Exit(1)
	}
}
