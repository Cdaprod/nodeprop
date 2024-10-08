// cmd/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cdaprod/nodeprop/pkg/nodeprop" // Correct import path
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// SignalHandler defines the structure for receiving signals to trigger actions.
type SignalHandler struct {
	SignalCh chan os.Signal
	ActionCh chan string
}

// NewSignalHandler initializes and returns a SignalHandler.
func NewSignalHandler() *SignalHandler {
	return &SignalHandler{
		SignalCh: make(chan os.Signal, 1),
		ActionCh: make(chan string, 1),
	}
}

// ListenForSignal waits for system signals and passes corresponding actions to the action channel.
func (sh *SignalHandler) ListenForSignal() {
	signal.Notify(sh.SignalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case sig := <-sh.SignalCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("\nReceived shutdown signal")
				sh.ActionCh <- "shutdown"
				return
			case syscall.SIGHUP:
				fmt.Println("\nReceived reload signal")
				sh.ActionCh <- "reload"
			}
		}
	}
}

// DynamicRunner is a generic function that runs any given function concurrently with dynamic arguments.
func DynamicRunner[T any](fn func(T) error, arg T, logger *logrus.Logger) {
	go func() {
		err := fn(arg)
		if err != nil {
			logger.Errorf("Error executing function: %v", err)
		} else {
			logger.Infof("Function executed successfully with arg: %+v", arg)
		}
	}()
}

// handleArgsOrSignals dynamically processes actions with generic argument types.
func handleArgsOrSignals[T any](np *nodeprop.NodePropManager, action string, arg T, logger *logrus.Logger) {
	switch action {
	case "add_workflow":
		// Add workflow using dynamic arguments passed via CLI or signal
		DynamicRunner(np.AddWorkflow, arg, logger) // Run in a Go routine
	case "shutdown":
		logger.Info("Shutting down NodePropManager...")
		DynamicRunner(func(_ T) error {
			time.Sleep(1 * time.Second) // Simulate some work
			fmt.Println("NodePropManager shutdown complete")
			return nil
		}, arg, logger) // Use empty struct or appropriate type as no argument is needed
	case "reload":
		logger.Info("Reloading configuration...")
		DynamicRunner(np.ReloadConfig, arg, logger)
	default:
		logger.Warnf("Unknown action: %s", action)
	}
}

// NodePropArguments holds dynamic arguments for generic actions.
type NodePropArguments struct {
	RepoPath string
	Workflow string
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Define CLI flags
	addWorkflow := flag.Bool("add-workflow", false, "Flag to add a new workflow")
	repoPath := flag.String("repo", "", "Path to the target repository")
	workflowName := flag.String("workflow", "", "Name of the workflow to add")
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Initialize Viper for configuration management
	viper.SetConfigFile(*configPath)
	viper.SetConfigType("yaml")

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		logger.Fatalf("Error reading config file: %v", err)
	}

	// Initialize NodePropManager with configuration
	np, err := nodeprop.NewNodePropManager(viper.GetString("global_nodeprop_path"), viper.GetString("workflow_template_path"), logger)
	if err != nil {
		logger.Fatalf("Failed to initialize NodePropManager: %v", err)
	}

	// Subscribe to events (if any)
	eventCh := np.SubscribeEvents()
	go func() {
		for event := range eventCh {
			switch event.Type {
			case nodeprop.EventTypeSuccess:
				logger.Infof("SUCCESS: %s", event.Message)
			case nodeprop.EventTypeError:
				logger.Errorf("ERROR: %s", event.Message)
			case nodeprop.EventTypeInfo:
				logger.Infof("INFO: %s", event.Message)
			}
		}
	}()

	// Initialize the signal handler
	signalHandler := NewSignalHandler()

	// Listen for system signals (like SIGINT, SIGTERM) in a separate goroutine
	go signalHandler.ListenForSignal()

	// Define dynamic arguments for adding a workflow
	args := nodeprop.NodePropArguments{
		RepoPath: *repoPath,
		Workflow: *workflowName,
	}

	// Handle CLI args or signal-based actions dynamically using generics
	go func() {
		if *addWorkflow {
			handleArgsOrSignals[np.AddWorkflow]("add_workflow", args, logger)
		}

		// Process actions from signals dynamically
		for action := range signalHandler.ActionCh {
			// For actions like "shutdown" or "reload", use appropriate argument types
			switch action {
			case "shutdown":
				var emptyArg struct{}
				handleArgsOrSignals[np.Shutdown](action, emptyArg, logger)
			case "reload":
				handleArgsOrSignals[np.ReloadConfig]("reload", args, logger)
			default:
				logger.Warnf("Unhandled action: %s", action)
			}
		}
	}()

	// Wait indefinitely until a shutdown signal is received
	select {}
}