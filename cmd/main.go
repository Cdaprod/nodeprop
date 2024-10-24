// cmd/main.go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/Cdaprod/nodeprop/cmd/cli"
    "github.com/Cdaprod/nodeprop/cmd/tui"
    "github.com/Cdaprod/nodeprop/pkg/nodeprop"
)

func main() {
    ctx := context.Background()

    // Initialize NodeProp manager
    manager, err := nodeprop.New(
        nodeprop.WithGitHubToken(os.Getenv("GITHUB_TOKEN")),
        nodeprop.WithLogger(nodeprop.NewDefaultLogger()),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error initializing nodeprop: %v\n", err)
        os.Exit(1)
    }

    // Check if TUI mode is requested
    if len(os.Args) > 1 && os.Args[1] == "--tui" {
        if err := tui.Run(ctx, manager); err != nil {
            fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
            os.Exit(1)
        }
        return
    }

    // Default to CLI mode
    if err := cli.Execute(ctx, manager); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}