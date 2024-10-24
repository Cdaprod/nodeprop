// pkg/nodeprop/interfaces.go
package nodeprop

import (
    "context"
    "time"
)

// CoreManager combines all manager interfaces
type CoreManager interface {
    ConfigManager
    WorkflowManager
    SecretManager
    RepositoryManager
    EventEmitter
}

// ConfigManager handles configuration related operations
type ConfigManager interface {
    LoadConfig(ctx context.Context) error
    SaveConfig(ctx context.Context) error
    GetConfigValue(key string) interface{}
    SetConfigValue(key string, value interface{}) error
}

// WorkflowManager handles GitHub workflow operations
type WorkflowManager interface {
    AddWorkflow(ctx context.Context, args WorkflowArguments) error
    UpdateWorkflow(ctx context.Context, args WorkflowArguments) error
    DeleteWorkflow(ctx context.Context, repo, name string) error
    ListWorkflows(ctx context.Context, repo string) ([]Workflow, error)
    TriggerWorkflow(ctx context.Context, repo, workflowID string, inputs map[string]interface{}) error
}

// SecretManager handles GitHub secrets
type SecretManager interface {
    AddSecret(ctx context.Context, args SecretArguments) error
    DeleteSecret(ctx context.Context, repo, name string) error
    ListSecrets(ctx context.Context, repo string) ([]Secret, error)
}

// RepositoryManager handles repository operations
type RepositoryManager interface {
    GenerateNodeProp(ctx context.Context, args NodePropArguments) error
    UpdateNodeProp(ctx context.Context, args NodePropArguments) error
    ValidateNodeProp(ctx context.Context, nodeProp NodePropFile) error
    CheckFile(ctx context.Context, repo, path string) (bool, []byte, error)
}

// EventEmitter handles event publishing and subscription
type EventEmitter interface {
    Subscribe(eventType EventType) (<-chan Event, func())
    Emit(event Event)
}

// Store interface for persistent storage
type Store interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
    List(prefix string) (map[string]interface{}, error)
}

// Cache interface for temporary storage
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, expiration time.Duration)
    Delete(key string)
}

// Arguments structures
type WorkflowArguments struct {
    Repository string
    Name       string
    Content    string
    Template   string
    Variables  map[string]interface{}
    Reference  string    // For triggering workflows
}

type SecretArguments struct {
    Repository string
    Name       string
    Value      string
    Visibility string // "all", "private", "selected"
}

type NodePropArguments struct {
    RepoPath  string
    RepoName  string
    Domain    string
    Config    string
    Variables map[string]interface{}
}

// Result structures
type Workflow struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Path     string    `json:"path"`
    Content  string    `json:"content"`
    Created  time.Time `json:"created"`
    Updated  time.Time `json:"updated"`
    Status   string    `json:"status"`
}

type Secret struct {
    Name       string    `json:"name"`
    Created    time.Time `json:"created"`
    Updated    time.Time `json:"updated"`
    Visibility string    `json:"visibility"`
}

// Event types and structures
type EventType string

const (
    EventWorkflow EventType = "workflow"
    EventSecret   EventType = "secret"
    EventConfig   EventType = "config"
    EventError    EventType = "error"
)

type Event struct {
    Type      EventType              `json:"type"`
    Name      string                 `json:"name"`
    Data      interface{}            `json:"data"`
    Error     error                  `json:"error,omitempty"`
    Timestamp time.Time             `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Factory method and options
type Option func(*NodePropManager) error

func NewNodePropManager(ctx context.Context, opts ...Option) (*NodePropManager, error) {
    manager := &NodePropManager{
        config:   DefaultConfig(),
        store:    NewFileStore(),
        cache:    NewInMemoryCache(),
        eventBus: NewEventBus(),
        logger:   NewLogger(),
    }

    for _, opt := range opts {
        if err := opt(manager); err != nil {
            return nil, err
        }
    }

    if err := manager.Initialize(ctx); err != nil {
        return nil, err
    }

    return manager, nil
}

// Configuration options
func WithGitHubToken(token string) Option {
    return func(m *NodePropManager) error {
        m.config.GitHub.Token = token
        return nil
    }
}

func WithLogger(logger Logger) Option {
    return func(m *NodePropManager) error {
        m.logger = logger
        return nil
    }
}

func WithStore(store Store) Option {
    return func(m *NodePropManager) error {
        m.store = store
        return nil
    }
}

func WithCache(cache Cache) Option {
    return func(m *NodePropManager) error {
        m.cache = cache
        return nil
    }
}

// Logger interface
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    WithField(key string, value interface{}) Logger
}