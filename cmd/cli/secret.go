 cmd/nodeprop/cli/secret.go
package cli

import (
    "github.com/spf13/cobra"
)

func newSecretCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "secret",
        Short: "Manage GitHub secrets",
        Long:  `Add and manage repository secrets securely.`,
    }

    cmd.AddCommand(
        newSecretAddCmd(),
        newSecretListCmd(),
    )

    return cmd
}

func newSecretAddCmd() *cobra.Command {
    var (
        name  string
        value string
        repo  string
    )

    cmd := &cobra.Command{
        Use:   "add",
        Short: "Add a new secret",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            return manager.AddSecret(ctx, nodeprop.SecretArguments{
                Name:  name,
                Value: value,
                Repository: repo,
            })
        },
    }

    cmd.Flags().StringVarP(&name, "name", "n", "", "secret name")
    cmd.Flags().StringVarP(&value, "value", "v", "", "secret value")
    cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository (owner/repo)")
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("repo")

    return cmd
}
