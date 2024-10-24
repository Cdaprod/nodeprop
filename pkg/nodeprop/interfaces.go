// pkg/nodeprop/interfaces.go
package nodeprop

import (
    "context"
)

// Core interfaces
type Manager interface {
    ConfigManager
    WorkflowManager
    SecretManager
    EventEmitter
}

type ConfigManager interface {
    AddNodePropConfig(ctx context.Context, args NodePropArguments) error
    ReloadConfig(args NodePropArguments) error
}

type WorkflowManager interface {
    AddWorkflow(ctx context.Context, args NodePropArguments) error
    ListWorkflows(ctx context.Context) ([]string, error)
}

type SecretManager interface {
    AddSecret(ctx context.Context, repo, name, value string) error
    ListSecrets(ctx context.Context, repo string) ([]string, error)
}

type EventEmitter interface {
    Subscribe(eventType EventType) <-chan Event
    Emit(event Event)
}

// Enhanced types
type NodePropManager struct {
    config      *Config
    logger      *logrus.Logger
    secretMgr   SecretManager
    workflowMgr WorkflowManager
    eventBus    EventEmitter
    store       Store
}

type Config struct {
    GlobalNodePropPath   string            `yaml:"global_nodeprop_path"`
    WorkflowTemplatePath string            `yaml:"workflow_template_path"`
    GitHub              GitHubConfig       `yaml:"github"`
    Storage            StorageConfig       `yaml:"storage"`
}

type Store interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    Delete(key string) error
}

// Factory function
func NewNodePropManager(ctx context.Context, opts ...Option) (*NodePropManager, error) {
    npm := &NodePropManager{
        config:   DefaultConfig(),
        eventBus: NewEventBus(),
        store:    NewFileStore(),
    }

    for _, opt := range opts {
        if err := opt(npm); err != nil {
            return nil, err
        }
    }

    return npm, nil
}

// Options pattern
type Option func(*NodePropManager) error

func WithGitHub(token string) Option {
    return func(npm *NodePropManager) error {
        secretMgr, err := NewGitHubSecretManager(token)
        if err != nil {
            return err
        }
        npm.secretMgr = secretMgr
        return nil
    }
}

func WithLogger(logger *logrus.Logger) Option {
    return func(npm *NodePropManager) error {
        npm.logger = logger
        return nil
    }
}

// Event system
type EventBus struct {
    subscribers map[EventType][]chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType EventType) <-chan Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    ch := make(chan Event, 100)
    eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
    return ch
}

func (eb *EventBus) Emit(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()

    for _, ch := range eb.subscribers[event.Type] {
        select {
        case ch <- event:
        default:
            // Channel full, skip
        }
    }
}

// Example usage
func ExampleUsage() {
    ctx := context.Background()
    logger := logrus.New()

    npm, err := NewNodePropManager(
        ctx,
        WithGitHub(os.Getenv("GITHUB_TOKEN")),
        WithLogger(logger),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Subscribe to events
    events := npm.eventBus.Subscribe(EventTypeWorkflow)
    go func() {
        for event := range events {
            logger.Infof("Workflow event: %v", event)
        }
    }()

    // Use the manager
    err = npm.AddWorkflow(ctx, NodePropArguments{
        RepoPath: "./myrepo",
        Workflow: "ci.yml",
    })
}