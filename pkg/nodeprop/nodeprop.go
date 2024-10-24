// pkg/nodeprop/nodeprop.go
package nodeprop

import (
    "context"
    "github.com/sirupsen/logrus"
)

// Public interface
type Service interface {
    AddWorkflow(ctx context.Context, args Arguments) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Subscribe(eventType EventType) <-chan Event
}

// Public configuration
type Config struct {
    CacheSize      int    `yaml:"cache_size"`
    MetricsEnabled bool   `yaml:"metrics_enabled"`
    AssetsDir      string `yaml:"assets_dir"`
}

// Public factory method
func NewService(logger *logrus.Logger, config Config) (Service, error) {
    return InitializeNodePropService(logger, config)
}

// examples/basic/main.go
package main

import (
    "context"
    "log"

    "github.com/Cdaprod/nodeprop"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    config := nodeprop.Config{
        CacheSize:      1000,
        MetricsEnabled: true,
        AssetsDir:      "./assets",
    }

    service, err := nodeprop.NewService(logger, config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := service.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer service.Stop(ctx)

    args := nodeprop.Arguments{
        RepoPath: "./myrepo",
        Workflow: "test-workflow",
        Domain:   "example.com",
    }

    if err := service.AddWorkflow(ctx, args); err != nil {
        log.Fatal(err)
    }
}