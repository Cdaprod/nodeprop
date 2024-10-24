// cmd/nodeprop/cli/root.go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/Cdaprod/nodeprop/pkg/nodeprop"
)

var (
    cfgFile string
    verbose bool
    manager *nodeprop.NodePropManager
)

// rootCmd represents the base command
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "nodeprop",
        Short: "NodeProp - Repository Configuration Management",
        Long: `NodeProp is a tool for managing repository configurations,
GitHub workflows, and secrets programmatically.`,
        PersistentPreRunE: initializeManager,
    }

    // Global flags
    cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
    cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

    // Add commands
    cmd.AddCommand(
        newWorkflowCmd(),
        newSecretCmd(),
        newCheckCmd(),
        newConfigCmd(),
    )

    return cmd
}

// Execute runs the root command
func Execute() error {
    return NewRootCmd().Execute()
}

// cmd/nodeprop/cli/workflow.go
package cli

import (
    "github.com/spf13/cobra"
)

func newWorkflowCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "workflow",
        Short: "Manage GitHub workflows",
        Long:  `Add, update, and trigger GitHub workflows.`,
    }

    // Add subcommands
    cmd.AddCommand(
        newWorkflowAddCmd(),
        newWorkflowTriggerCmd(),
        newWorkflowStatusCmd(),
    )

    return cmd
}

func newWorkflowAddCmd() *cobra.Command {
    var (
        name     string
        template string
        repo     string
    )

    cmd := &cobra.Command{
        Use:   "add",
        Short: "Add a new workflow",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            return manager.AddWorkflow(ctx, nodeprop.WorkflowArguments{
                Name:     name,
                Template: template,
                Repository: repo,
            })
        },
    }

    cmd.Flags().StringVarP(&name, "name", "n", "", "workflow name")
    cmd.Flags().StringVarP(&template, "template", "t", "", "workflow template")
    cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository (owner/repo)")
    cmd.MarkFlagRequired("name")
    cmd.MarkFlagRequired("repo")

    return cmd
}