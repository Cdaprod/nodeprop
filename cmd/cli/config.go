// cmd/nodeprop/cli/config.go
package cli

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "gopkg.in/yaml.v2"
)

func newConfigCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "config",
        Short: "Manage NodeProp configuration",
        Long: `View and modify NodeProp configuration settings.
Handles both global configuration and repository-specific settings.`,
    }

    cmd.AddCommand(
        newConfigViewCmd(),
        newConfigSetCmd(),
        newConfigInitCmd(),
        newConfigValidateCmd(),
        newConfigImportCmd(),
        newConfigExportCmd(),
    )

    return cmd
}

func newConfigViewCmd() *cobra.Command {
    var (
        format string
        repo   string
    )

    cmd := &cobra.Command{
        Use:   "view [key]",
        Short: "View configuration settings",
        Long: `View current configuration settings.
Optionally specify a key to view specific settings.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            var config interface{}
            
            if repo != "" {
                // Get repository-specific configuration
                ctx := cmd.Context()
                repoConfig, err := manager.GetRepositoryConfig(ctx, repo)
                if err != nil {
                    return err
                }
                config = repoConfig
            } else if len(args) > 0 {
                // Get specific configuration key
                config = viper.Get(args[0])
            } else {
                // Get all settings
                config = viper.AllSettings()
            }

            // Format output
            var output []byte
            var err error

            switch strings.ToLower(format) {
            case "json":
                output, err = json.MarshalIndent(config, "", "  ")
            case "yaml":
                output, err = yaml.Marshal(config)
            default:
                return fmt.Errorf("unsupported format: %s", format)
            }

            if err != nil {
                return fmt.Errorf("error formatting config: %w", err)
            }

            fmt.Println(string(output))
            return nil
        },
    }

    cmd.Flags().StringVarP(&format, "format", "f", "yaml", "output format (json or yaml)")
    cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository-specific config (owner/repo)")

    return cmd
}

func newConfigSetCmd() *cobra.Command {
    var (
        repo string
    )

    cmd := &cobra.Command{
        Use:   "set [key] [value]",
        Short: "Set configuration value",
        Args:  cobra.ExactArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            key, value := args[0], args[1]

            if repo != "" {
                // Set repository-specific configuration
                ctx := cmd.Context()
                return manager.SetRepositoryConfig(ctx, repo, key, value)
            }

            // Set global configuration
            viper.Set(key, value)
            if err := viper.WriteConfig(); err != nil {
                return fmt.Errorf("error writing config: %w", err)
            }

            fmt.Printf("Successfully set %s = %s\n", key, value)
            return nil
        },
    }

    cmd.Flags().StringVarP(&repo, "repo", "r", "", "repository-specific config (owner/repo)")

    return cmd
}

func newConfigInitCmd() *cobra.Command {
    var (
        force bool
    )

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize default configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            configFile := viper.ConfigFileUsed()
            if configFile != "" && !force {
                return fmt.Errorf("configuration file already exists at %s. Use --force to overwrite", configFile)
            }

            // Default configuration
            defaultConfig := map[string]interface{}{
                "github": map[string]interface{}{
                    "default_branch": "main",
                    "default_visibility": "private",
                },
                "workflows": map[string]interface{}{
                    "template_dir": "templates/workflows",
                    "default_template": "default.yml",
                },
                "nodeprop": map[string]interface{}{
                    "template_path": "templates/nodeprop.yml",
                    "auto_generate": true,
                },
                "cache": map[string]interface{}{
                    "enabled": true,
                    "ttl": "1h",
                },
            }

            // Write configuration
            for k, v := range defaultConfig {
                viper.Set(k, v)
            }

            if err := viper.SafeWriteConfig(); err != nil {
                if err := viper.WriteConfig(); err != nil {
                    return fmt.Errorf("error writing config: %w", err)
                }
            }

            fmt.Printf("Configuration initialized at %s\n", viper.ConfigFileUsed())
            return nil
        },
    }

    cmd.Flags().BoolVarP(&force, "force", "f", false, "force overwrite existing configuration")

    return cmd
}

func newConfigValidateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "validate",
        Short: "Validate current configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Validate global configuration
            if err := manager.ValidateConfig(); err != nil {
                return fmt.Errorf("configuration validation failed: %w", err)
            }

            fmt.Println("Configuration is valid")
            return nil
        },
    }
}

func newConfigImportCmd() *cobra.Command {
    var (
        format string
    )

    cmd := &cobra.Command{
        Use:   "import [file]",
        Short: "Import configuration from file",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            data, err := os.ReadFile(args[0])
            if err != nil {
                return fmt.Errorf("error reading file: %w", err)
            }

            var config map[string]interface{}
            switch strings.ToLower(format) {
            case "json":
                if err := json.Unmarshal(data, &config); err != nil {
                    return fmt.Errorf("error parsing JSON: %w", err)
                }
            case "yaml":
                if err := yaml.Unmarshal(data, &config); err != nil {
                    return fmt.Errorf("error parsing YAML: %w", err)
                }
            default:
                return fmt.Errorf("unsupported format: %s", format)
            }

            // Update configuration
            for k, v := range config {
                viper.Set(k, v)
            }

            if err := viper.WriteConfig(); err != nil {
                return fmt.Errorf("error writing config: %w", err)
            }

            fmt.Println("Configuration imported successfully")
            return nil
        },
    }

    cmd.Flags().StringVarP(&format, "format", "f", "yaml", "input format (json or yaml)")

    return cmd
}

func newConfigExportCmd() *cobra.Command {
    var (
        format string
        output string
    )

    cmd := &cobra.Command{
        Use:   "export",
        Short: "Export configuration to file",
        RunE: func(cmd *cobra.Command, args []string) error {
            config := viper.AllSettings()

            // Format configuration
            var data []byte
            var err error

            switch strings.ToLower(format) {
            case "json":
                data, err = json.MarshalIndent(config, "", "  ")
            case "yaml":
                data, err = yaml.Marshal(config)
            default:
                return fmt.Errorf("unsupported format: %s", format)
            }

            if err != nil {
                return fmt.Errorf("error formatting config: %w", err)
            }

            // Write to file or stdout
            if output == "-" {
                fmt.Println(string(data))
            } else {
                if err := os.WriteFile(output, data, 0644); err != nil {
                    return fmt.Errorf("error writing file: %w", err)
                }
                fmt.Printf("Configuration exported to %s\n", output)
            }

            return nil
        },
    }

    cmd.Flags().StringVarP(&format, "format", "f", "yaml", "output format (json or yaml)")
    cmd.Flags().StringVarP(&output, "output", "o", "-", "output file (- for stdout)")

    return cmd
}