[![CI](https://github.com/Cdaprod/nodeprop/actions/workflows/ci.yml/badge.svg)](https://github.com/Cdaprod/nodeprop/actions/workflows/ci.yml)

# NodeProp

NodeProp is a dynamic workflow management system designed to automate the addition and management of workflows within target repositories. It leverages Go's powerful features, including generics and Viper for configuration management, to provide a flexible and scalable solution for managing workflows and their configurations.

## Features

- **Dynamic Workflow Addition**: Easily add new workflows to target repositories.
- **Configuration Management**: Utilize Viper for managing and reloading configurations.
- **Signal Handling**: Gracefully handle system signals for shutdowns and configuration reloads.
- **Generics for Flexibility**: Use Go's generics to handle various actions and arguments dynamically.
- **Automated Configuration File Generation**: Automatically generate and manage `.nodeprop.yml` configuration files based on workflows.

## Getting Started

### Prerequisites

- Go 1.20 or higher
- Git
- GitHub Actions (for CI workflows)
- YAML support

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/Cdaprod/nodeprop.git
   cd nodeprop

2.	Install Dependencies

go mod tidy


3.	Setup Configuration
Create a config.yaml file in the root directory with the following content:

global_nodeprop_path: "./assets/.empty.nodeprop.yml"
workflow_template_path: "./assets/index-nodeprop-workflow.yml"

Ensure that the assets directory contains .empty.nodeprop.yml and index-nodeprop-workflow.yml templates.

### Usage

#### Adding a Workflow

To add a new workflow to a target repository and generate a .nodeprop.yml file:

go run cmd/main.go --add-workflow --repo /path/to/repo --workflow new-workflow --domain your.domain --config ./config.yaml

	•	--add-workflow: Flag to trigger the addition of a new workflow.
	•	--repo: Path to the target repository.
	•	--workflow: Name of the workflow to add.
	•	--domain: Domain under which the service is registered.
	•	--config: Path to the configuration file.

#### Handling Signals

The application listens for system signals such as SIGINT, SIGTERM, and SIGHUP to perform actions like shutdown and configuration reloads.

	•	Shutdown: Send SIGINT or SIGTERM to gracefully shut down the application.
	•	Reload Configuration: Send SIGHUP to reload the configuration file without restarting the application.

### Testing

NodeProp includes comprehensive tests to ensure functionality.

go test ./pkg/nodeprop/...

### CI/CD

NodeProp uses GitHub Actions for continuous integration. The CI workflow is defined in .github/workflows/ci.yml.

#### Project Structure

```
nodeprop/
├── cmd/
│   └── main.go                 // CLI entry point
├── pkg/
│   └── nodeprop/
│       ├── manager.go          // Core logic for managing workflows and nodeprop files
│       ├── manager_test.go     // Tests for NodePropManager
│       ├── types.go            // Definitions of structures like NodePropFile, Metadata, etc.
│       ├── config.go           // Configuration management using Viper
│       └── utils.go            // Utility functions
├── assets/
│   ├── .empty.nodeprop.yml     // Template for the .nodeprop.yml file
│   └── index-nodeprop-workflow.yml // Template for GitHub Actions workflow
├── .github/
│   └── workflows/
│       └── ci.yml              // GitHub Actions CI workflow
├── config.yaml                 // Configuration file
├── go.mod                      // Go module dependencies
└── go.sum    // Dependency checksum file
``` 

## Contributing

Contributions are welcome! Please follow these steps:

	1.	Fork the repository.
	2.	Create your feature branch (git checkout -b feature/NewFeature).
	3.	Commit your changes (git commit -m 'Add some feature').
	4.	Push to the branch (git push origin feature/NewFeature).
	5.	Open a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contact

For any inquiries or support, please contact Cdaprod.
