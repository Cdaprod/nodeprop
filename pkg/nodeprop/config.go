// pkg/nodeprop/config.go
package nodeprop

import (
	"fmt"

	"github.com/spf13/viper"
)

// NodePropManager handles adding workflows and managing .nodeprop.yml files
type NodePropManager struct {
	GlobalNodePropPath string
	WorkflowTemplatePath string
	Logger             *logrus.Logger
}

// NewNodePropManager initializes the NodePropManager with paths from the config
func NewNodePropManager(globalNodePropPath, workflowTemplatePath string, logger *logrus.Logger) (*NodePropManager, error) {
	if globalNodePropPath == "" {
		return nil, fmt.Errorf("global_nodeprop_path is required")
	}
	if workflowTemplatePath == "" {
		return nil, fmt.Errorf("workflow_template_path is required")
	}

	return &NodePropManager{
		GlobalNodePropPath: globalNodePropPath,
		WorkflowTemplatePath: workflowTemplatePath,
		Logger:             logger,
	}, nil
}

// ReloadConfig reloads the Viper configuration
func (npm *NodePropManager) ReloadConfig(arg NodePropArguments) error {
	err := viper.ReadInConfig()
	if err != nil {
		npm.Logger.Errorf("Error reading config file: %v", err)
		return err
	}
	npm.Logger.Info("Configuration reloaded successfully")
	return nil
}