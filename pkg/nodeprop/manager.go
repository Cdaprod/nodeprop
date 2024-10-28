// pkg/nodeprop/manager.go
package nodeprop

import (
	"fmt"
	"io/ioutil"
	"os"
//	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"github.com/spf13/viper"
	"os/signal"
	"syscall"
)

// NodePropManager represents the manager handling node properties and workflows.
type NodePropManager struct {
	Logger *logrus.Logger
}

// EventType represents the type of an event (e.g., success, error, info).
type EventType string

const (
	EventTypeSuccess EventType = "success"
	EventTypeError   EventType = "error"
	EventTypeInfo    EventType = "info"
)

// Event represents a system event with type and message.
type Event struct {
	Type    EventType
	Message string
}

// NodePropArguments holds the arguments required for a NodeProp operation.
type NodePropArguments struct {
	RepoPath  string
	Workflow  string
	Domain    string
	Config    string
}

// NodePropFile represents the structure of a generated .nodeprop.yml file.
type NodePropFile struct {
	ID               string            `yaml:"id"`
	Name             string            `yaml:"name"`
	Address          string            `yaml:"address"`
	Capabilities     []string          `yaml:"capabilities"`
	Status           string            `yaml:"status"`
	Metadata         Metadata          `yaml:"metadata"`
	CustomProperties CustomProperties  `yaml:"custom_properties"`
}

// AddWorkflow adds a new workflow to the target repository using `index-nodeprop-workflow.yml` 
// and generates `.nodeprop.yml` using a template from `/assets/.empty.nodeprop.yml`.
func (npm *NodePropManager) AddWorkflow(args NodePropArguments) error {
	npm.Logger.Infof("Adding workflow '%s' to repository '%s'", args.Workflow, args.RepoPath)

	// Path to the local assets folder containing the workflow and .empty.nodeprop.yml.
	assetsDir := "./assets"

	// Read the `index-nodeprop-workflow.yml` from assets directory.
	workflowFile := filepath.Join(assetsDir, "index-nodeprop-workflow.yml")
	workflowContent, err := ioutil.ReadFile(workflowFile)
	if err != nil {
		npm.Logger.Errorf("Failed to read workflow file '%s': %v", workflowFile, err)
		return err
	}

	// Write the workflow to the target repo's `.github/workflows` directory.
	workflowPath := filepath.Join(args.RepoPath, ".github", "workflows", fmt.Sprintf("%s.yml", args.Workflow))
	err = os.MkdirAll(filepath.Dir(workflowPath), 0755)
	if err != nil {
		npm.Logger.Errorf("Failed to create workflow directory: %v", err)
		return err
	}

	err = ioutil.WriteFile(workflowPath, workflowContent, 0644)
	if err != nil {
		npm.Logger.Errorf("Failed to write workflow file: %v", err)
		return err
	}

	npm.Logger.Infof("Workflow '%s' added successfully to repository '%s'", args.Workflow, args.RepoPath)

	// Simulate workflow execution and generating `.nodeprop.yml`.
	npm.Logger.Info("Waiting for workflow to complete...")
	time.Sleep(5 * time.Second) // Simulated delay.

	// Read the `.empty.nodeprop.yml` template from assets directory.
	emptyNodePropFile := filepath.Join(assetsDir, ".empty.nodeprop.yml")
	emptyNodePropContent, err := ioutil.ReadFile(emptyNodePropFile)
	if err != nil {
		npm.Logger.Errorf("Failed to read .empty.nodeprop.yml: %v", err)
		return err
	}

	// Unmarshal the empty nodeprop template.
	var nodeProp NodePropFile
	err = yaml.Unmarshal(emptyNodePropContent, &nodeProp)
	if err != nil {
		npm.Logger.Errorf("Failed to unmarshal .empty.nodeprop.yml: %v", err)
		return err
	}

	// Update the nodeprop template with dynamic values.
	nodeProp.ID = uuid.New().String()
	nodeProp.Name = filepath.Base(args.RepoPath)
	nodeProp.Address = fmt.Sprintf("https://github.com/Cdaprod/%s", filepath.Base(args.RepoPath))
	nodeProp.Metadata.LastUpdated = time.Now().Format(time.RFC3339)
	nodeProp.CustomProperties.Domain = args.Domain

	// Marshal the updated .nodeprop.yml file.
	nodePropYAML, err := yaml.Marshal(&nodeProp)
	if err != nil {
		npm.Logger.Errorf("Failed to marshal .nodeprop.yml: %v", err)
		return err
	}

	// Write the updated .nodeprop.yml to the target repository.
	nodePropPath := filepath.Join(args.RepoPath, ".nodeprop.yml")
	err = ioutil.WriteFile(nodePropPath, nodePropYAML, 0644)
	if err != nil {
		npm.Logger.Errorf("Failed to write .nodeprop.yml: %v", err)
		return err
	}

	npm.Logger.Infof(".nodeprop.yml generated successfully at %s", nodePropPath)
	return nil
}

// SignalHandler listens for OS signals to handle reloads or shutdowns.
func (npm *NodePropManager) SignalHandler() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		sig := <-signalCh
		switch sig {
		case syscall.SIGHUP:
			npm.Logger.Info("Received SIGHUP, reloading configuration.")
			npm.ReloadConfig(NodePropArguments{Config: "config.yaml"})
		case syscall.SIGINT, syscall.SIGTERM:
			npm.Logger.Info("Received termination signal, shutting down.")
			os.Exit(0)
		}
	}
}

// ReloadConfig reloads the configuration using Viper.
func (npm *NodePropManager) ReloadConfig(args NodePropArguments) error {
	viper.SetConfigFile(args.Config) // Use the specified config file.
	err := viper.ReadInConfig()
	if err != nil {
		npm.Logger.Errorf("Error reading config file during reload: %v", err)
		return err
	}
	npm.Logger.Info("Configuration reloaded successfully.")
	return nil
}