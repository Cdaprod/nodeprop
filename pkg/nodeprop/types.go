// pkg/nodeprop/types.go
package nodeprop

// NodePropFile represents the structure of the .nodeprop.yml file
type NodePropFile struct {
	ID               string           `yaml:"id"`
	Name             string           `yaml:"name"`
	Address          string           `yaml:"address"`
	Capabilities     []string         `yaml:"capabilities"`
	Status           string           `yaml:"status"`
	Metadata         Metadata         `yaml:"metadata"`
	CustomProperties CustomProperties  `yaml:"custom_properties"`
}

// Metadata defines the metadata section in .nodeprop.yml
type Metadata struct {
	Description string `yaml:"description"`
	Owner       string `yaml:"owner"`
	LastUpdated string `yaml:"last_updated"`
	Tags        []string `yaml:"tags"`
	GitHub      GitHub   `yaml:"github"`
	Docker      Docker   `yaml:"docker"`
}

// GitHub metadata about the repository.
type GitHub struct {
	Stars        int    `yaml:"stars"`
	Forks        int    `yaml:"forks"`
	Issues       int    `yaml:"issues"`
	PullRequests PRInfo `yaml:"pull_requests"`
	LatestCommit string `yaml:"latest_commit"`
	License      string `yaml:"license"`
	Topics       []string `yaml:"topics"`
}

// PRInfo contains details about pull requests in the repository
type PRInfo struct {
	Open   int `yaml:"open"`
	Closed int `yaml:"closed"`
}

// Docker metadata for Docker containerization settings.
type Docker struct {
	Dockerfile    DockerfileInfo `yaml:"dockerfile"`
	DockerCompose DockerCompose  `yaml:"docker_compose"`
}

// DockerfileInfo stores Dockerfile data
type DockerfileInfo struct {
	ExposedPorts []string `yaml:"exposed_ports"`
	EnvVars      []string `yaml:"env_vars"`
	Cmd          string   `yaml:"cmd"`
	Entrypoint   string   `yaml:"entrypoint"`
	Volumes      []string `yaml:"volumes"`
}

// DockerCompose contains service-level docker-compose data
type DockerCompose struct {
	Services []Service `yaml:"services"`
	Ports    map[string][]int `yaml:"ports"`
	Volumes  map[string][]string `yaml:"volumes"`
	EnvVars  map[string][]string `yaml:"env_vars"`
	Command  map[string]string `yaml:"command"`
}

// Service defines an individual docker-compose service
type Service struct {
	Name    string   `yaml:"name"`
	Ports   []string `yaml:"ports"`
	EnvVars []string `yaml:"env_vars"`
	Volumes []string `yaml:"volumes"`
}

// CustomProperties represents custom fields in the nodeprop
type CustomProperties struct {
	DeployEnvironment string   `yaml:"deploy_environment"`
	MonitoringEnabled bool     `yaml:"monitoring_enabled"`
	AutoScale         bool     `yaml:"auto_scale"`
	Service           string   `yaml:"service"`
	App               string   `yaml:"app"`
	Image             string   `yaml:"image"`
	Ports             []string `yaml:"ports"`
	Volumes           []string `yaml:"volumes"`
	Network           string   `yaml:"network"`
	Domain            string   `yaml:"domain"`
}