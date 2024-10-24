// cmd/nodeprop/cli/check.go
package cli

import (
    "fmt"
    "github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
    var repo string

    cmd := &cobra.Command{
        Use:   "check [file]",
        Short: "Check for file existence",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            exists, content, err := manager.CheckFile(ctx, repo, args[0])
            if err != nil {
                return err
            }

            if exists {
                fmt.Printf("File '%s' exists in repository\n", args[0])
                if content != nil {
                    fmt.Println("Content:", content)
                }
            } else {
                fmt.Printf("File '%s' not found in repository\n", args[0])
            }
            return nil
        },
    }

    cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository (owner/repo)")
    cmd.MarkFlagRequired("repo")

    return cmd
}