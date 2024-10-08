// pkg/nodeprop/manager_test.go
package nodeprop

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// Helper function to create a temporary directory for testing
func setupTempRepo(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "nodeprop_test_repo")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir
}

// Helper function to clean up the temporary directory after testing
func teardownTempRepo(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to remove temp directory: %v", err)
	}
}

func TestAddWorkflow(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Setup temporary repository
	repoPath := setupTempRepo(t)
	defer teardownTempRepo(t, repoPath)

	// Setup assets directory with required files
	assetsDir := filepath.Join(repoPath, "assets")
	err := os.MkdirAll(assetsDir, 0755)
	assert.NoError(t, err, "Failed to create assets directory")

	// Create .empty.nodeprop.yml template
	emptyNodePropTemplate := `
id: ""
name: ""
address: ""
capabilities: []
status: ""
metadata:
  description: ""
  owner: ""
  last_updated: ""
  tags: []
  github:
    stars: 0
    forks: 0
    issues: 0
    pull_requests:
      open: 0
      closed: 0
    latest_commit: ""
    license: ""
    topics: []
  docker:
    dockerfile:
      exposed_ports: []
      env_vars: []
      cmd: ""
      entrypoint: ""
      volumes: []
    docker_compose:
      services: []
      ports: {}
      volumes: {}
      env_vars: {}
      command: {}
custom_properties:
  deploy_environment: null
  monitoring_enabled: false
  auto_scale: false
  service: ""
  app: ""
  image: ""
  ports: []
  volumes: []
  network: ""
  domain: ""
`
	err = ioutil.WriteFile(filepath.Join(assetsDir, ".empty.nodeprop.yml"), []byte(emptyNodePropTemplate), 0644)
	assert.NoError(t, err, "Failed to write .empty.nodeprop.yml")

	// Create index-nodeprop-workflow.yml template
	indexWorkflowTemplate := `
name: TestWorkflow

on:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v2
      - name: Run NodeProp Workflow
        run: echo "Running TestWorkflow"
`
	err = ioutil.WriteFile(filepath.Join(assetsDir, "index-nodeprop-workflow.yml"), []byte(indexWorkflowTemplate), 0644)
	assert.NoError(t, err, "Failed to write index-nodeprop-workflow.yml")

	// Initialize NodePropManager
	npManager := &NodePropManager{
		Logger: logger,
	}

	// Define NodePropArguments
	args := NodePropArguments{
		RepoPath: repoPath,
		Workflow: "test-workflow",
		Domain:   "test.domain",
		Config:   "config.yaml",
	}

	// Create a dummy config.yaml file
	configContent := `
global_nodeprop_path: "./assets/.empty.nodeprop.yml"
workflow_template_path: "./assets/index-nodeprop-workflow.yml"
`
	err = ioutil.WriteFile(filepath.Join(repoPath, "config.yaml"), []byte(configContent), 0644)
	assert.NoError(t, err, "Failed to write config.yaml")

	// Call AddWorkflow
	err = npManager.AddWorkflow(args)
	assert.NoError(t, err, "AddWorkflow failed")

	// Check if workflow file is created
	workflowPath := filepath.Join(repoPath, ".github", "workflows", "test-workflow.yml")
	_, err = os.Stat(workflowPath)
	assert.NoError(t, err, "Workflow file not created")

	// Check if .nodeprop.yml is generated
	nodePropPath := filepath.Join(repoPath, ".nodeprop.yml")
	_, err = os.Stat(nodePropPath)
	assert.NoError(t, err, ".nodeprop.yml file not created")

	// Read and verify .nodeprop.yml contents
	nodePropContent, err := ioutil.ReadFile(nodePropPath)
	assert.NoError(t, err, "Failed to read .nodeprop.yml")

	var nodeProp NodePropFile
	err = yaml.Unmarshal(nodePropContent, &nodeProp)
	assert.NoError(t, err, "Failed to unmarshal .nodeprop.yml")

	assert.NotEmpty(t, nodeProp.ID, "NodeProp ID should not be empty")
	assert.Equal(t, filepath.Base(repoPath), nodeProp.Name, "NodeProp Name mismatch")
	assert.Equal(t, fmt.Sprintf("https://github.com/Cdaprod/%s", filepath.Base(repoPath)), nodeProp.Address, "NodeProp Address mismatch")
	assert.Equal(t, "active", nodeProp.Status, "NodeProp Status should be active")
	assert.Equal(t, "test.domain", nodeProp.CustomProperties.Domain, "NodeProp Domain mismatch")
}

func TestReloadConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Setup temporary repository
	repoPath := setupTempRepo(t)
	defer teardownTempRepo(t, repoPath)

	// Create a config.yaml file
	configPath := filepath.Join(repoPath, "config.yaml")
	initialConfig := `
global_nodeprop_path: "./assets/.empty.nodeprop.yml"
workflow_template_path: "./assets/index-nodeprop-workflow.yml"
`
	err := ioutil.WriteFile(configPath, []byte(initialConfig), 0644)
	assert.NoError(t, err, "Failed to write initial config.yaml")

	// Initialize Viper and NodePropManager
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err, "Failed to read initial config.yaml")

	npManager := &NodePropManager{
		Logger: logger,
	}

	// Create a dummy new config.yaml
	newConfig := `
global_nodeprop_path: "./assets/.empty.nodeprop.yml"
workflow_template_path: "./assets/new_workflow_template.yml"
`
	err = ioutil.WriteFile(configPath, []byte(newConfig), 0644)
	assert.NoError(t, err, "Failed to write new config.yaml")

	// Call ReloadConfig
	args := NodePropArguments{
		Config: configPath,
	}
	err = npManager.ReloadConfig(args)
	assert.NoError(t, err, "ReloadConfig failed")

	// Verify the new configuration is loaded
	workflowTemplatePath := viper.GetString("workflow_template_path")
	assert.Equal(t, "./assets/new_workflow_template.yml", workflowTemplatePath, "Config reload did not update workflow_template_path correctly")
}