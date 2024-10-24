// pkg/nodeprop/manager.go
package nodeprop

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
)

// NodePropManager implements CoreManager interface
type NodePropManager struct {
    config      *Config
    logger      Logger
    store       Store
    cache       Cache
    eventBus    EventEmitter
    github      *GitHubClient
    validator   *Validator
    templates   *TemplateManager
    watcher     *ConfigWatcher
    mu          sync.RWMutex
}

// Initialize sets up the manager components
func (m *NodePropManager) Initialize(ctx context.Context) error {
    // Initialize GitHub client if token is provided
    if m.config.GitHub.Token != "" {
        client, err := NewGitHubClient(m.config.GitHub)
        if err != nil {
            return fmt.Errorf("failed to initialize GitHub client: %w", err)
        }
        m.github = client
    }

    // Start config watcher
    m.watcher.Start(ctx)

    // Load templates
    if err := m.templates.LoadTemplates(); err != nil {
        return fmt.Errorf("failed to load templates: %w", err)
    }

    return nil
}

// ConfigManager implementation

func (m *NodePropManager) LoadConfig(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Check cache first
    if cached, ok := m.cache.Get("config"); ok {
        if config, ok := cached.(*Config); ok {
            m.config = config
            return nil
        }
    }

    // Load from store
    data, err := m.store.Get("config")
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    config, ok := data.(*Config)
    if !ok {
        return fmt.Errorf("invalid config data type")
    }

    m.config = config
    m.cache.Set("config", config, 1*time.Hour)

    m.eventBus.Emit(Event{
        Type: EventTypeConfig,
        Name: "ConfigLoaded",
        Data: config,
    })

    return nil
}

func (m *NodePropManager) SaveConfig(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if err := m.store.Set("config", m.config); err != nil {
        return fmt.Errorf("failed to save config: %w", err)
    }

    m.cache.Set("config", m.config, 1*time.Hour)

    m.eventBus.Emit(Event{
        Type: EventTypeConfig,
        Name: "ConfigSaved",
        Data: m.config,
    })

    return nil
}

// WorkflowManager implementation

func (m *NodePropManager) AddWorkflow(ctx context.Context, args WorkflowArguments) error {
    // Validate arguments
    if err := m.validator.ValidateWorkflowArgs(args); err != nil {
        return fmt.Errorf("invalid workflow arguments: %w", err)
    }

    // Process template if specified
    content := args.Content
    if args.Template != "" {
        processed, err := m.templates.ProcessTemplate(args.Template, args.Variables)
        if err != nil {
            return fmt.Errorf("failed to process template: %w", err)
        }
        content = processed
    }

    // Add workflow using GitHub client
    if err := m.github.CreateWorkflow(ctx, args.Repository, args.Name, content); err != nil {
        return fmt.Errorf("failed to create workflow: %w", err)
    }

    // Update cache
    cacheKey := fmt.Sprintf("workflows:%s", args.Repository)
    m.cache.Delete(cacheKey)

    // Emit event
    m.eventBus.Emit(Event{
        Type: EventTypeWorkflow,
        Name: "WorkflowAdded",
        Data: Workflow{
            Name:    args.Name,
            Content: content,
            Updated: time.Now(),
            Status:  "active",
        },
    })

    return nil
}

// SecretManager implementation

func (m *NodePropManager) AddSecret(ctx context.Context, args SecretArguments) error {
    // Validate arguments
    if err := m.validator.ValidateSecretArgs(args); err != nil {
        return fmt.Errorf("invalid secret arguments: %w", err)
    }

    // Add secret using GitHub client
    if err := m.github.CreateSecret(ctx, args.Repository, args.Name, args.Value, args.Visibility); err != nil {
        return fmt.Errorf("failed to create secret: %w", err)
    }

    // Update cache
    cacheKey := fmt.Sprintf("secrets:%s", args.Repository)
    m.cache.Delete(cacheKey)

    // Store encrypted reference locally
    secretRef := SecretReference{
        Name:       args.Name,
        Repository: args.Repository,
        Created:    time.Now(),
        Updated:    time.Now(),
        Visibility: args.Visibility,
    }

    if err := m.store.Set(fmt.Sprintf("secret_refs:%s:%s", args.Repository, args.Name), secretRef); err != nil {
        m.logger.Warn("Failed to store secret reference locally")
    }

    // Emit event
    m.eventBus.Emit(Event{
        Type: EventTypeSecret,
        Name: "SecretAdded",
        Data: Secret{
            Name:       args.Name,
            Created:    time.Now(),
            Updated:    time.Now(),
            Visibility: args.Visibility,
        },
    })

    return nil
}

// RepositoryManager implementation

func (m *NodePropManager) GenerateNodeProp(ctx context.Context, args NodePropArguments) error {
    // Validate arguments
    if err := m.validator.ValidateNodePropArgs(args); err != nil {
        return fmt.Errorf("invalid nodeprop arguments: %w", err)
    }

    // Generate NodeProp configuration
    nodeProp := NodePropFile{
        ID:      uuid.New().String(),
        Name:    args.RepoName,
        Address: fmt.Sprintf("https://github.com/Cdaprod/%s", args.RepoName),
        Status:  "active",
        Metadata: Metadata{
            LastUpdated: time.Now().Format(time.RFC3339),
            Owner:      "Cdaprod",
        },
        CustomProperties: CustomProperties{
            Domain: args.Domain,
        },
    }

    // Validate generated config
    if err := m.validator.ValidateNodeProp(nodeProp); err != nil {
        return fmt.Errorf("invalid nodeprop configuration: %w", err)
    }

    // Convert to YAML
    data, err := yaml.Marshal(nodeProp)
    if err != nil {
        return fmt.Errorf("failed to marshal nodeprop: %w", err)
    }

    // Save to repository
    if err := m.github.CreateFile(ctx, args.RepoPath, ".nodeprop.yml", data); err != nil {
        return fmt.Errorf("failed to create nodeprop file: %w", err)
    }

    // Update cache
    cacheKey := fmt.Sprintf("nodeprop:%s", args.RepoPath)
    m.cache.Set(cacheKey, nodeProp, 1*time.Hour)

    // Emit event
    m.eventBus.Emit(Event{
        Type: EventTypeNodeProp,
        Name: "NodePropGenerated",
        Data: nodeProp,
    })

    return nil
}

// Shutdown handles graceful shutdown
func (m *NodePropManager) Shutdown(ctx context.Context) error {
    m.logger.Info("Shutting down NodePropManager...")

    // Stop config watcher
    m.watcher.Stop()

    // Flush cache
    if err := m.cache.(*Cache).Flush(); err != nil {
        m.logger.Warn("Failed to flush cache during shutdown")
    }

    // Save any pending configurations
    if err := m.SaveConfig(ctx); err != nil {
        m.logger.Warn("Failed to save config during shutdown")
    }

    m.eventBus.Emit(Event{
        Type: EventTypeSystem,
        Name: "Shutdown",
        Data: nil,
    })

    return nil
}