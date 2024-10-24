// pkg/nodeprop/github_ops.go
package nodeprop

import (
    "context"
    "encoding/base64"
    "fmt"
    "strings"
    "sync"

    "github.com/google/go-github/v53/github"
    "golang.org/x/oauth2"
)

// GitHubOperations handles direct GitHub API operations
type GitHubOperations struct {
    client     *github.Client
    logger     Logger
    cache      Cache
    encryptor  SecretEncryptor
    mu         sync.RWMutex
}

// SecretEncryptor handles GitHub secret encryption
type SecretEncryptor interface {
    Encrypt(value string, key *github.PublicKey) (string, error)
}

// NewGitHubOperations creates a new GitHub operations handler
func NewGitHubOperations(token string, logger Logger, cache Cache) *GitHubOperations {
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(ctx, ts)

    return &GitHubOperations{
        client:    github.NewClient(tc),
        logger:    logger,
        cache:     cache,
        encryptor: NewSecretEncryptor(),
    }
}

// Secret Operations

// AddSecret adds a secret to a repository
func (g *GitHubOperations) AddSecret(ctx context.Context, owner, repo, name, value string) error {
    // Get repository's public key for secret encryption
    pubKey, _, err := g.client.Actions.GetRepoPublicKey(ctx, owner, repo)
    if err != nil {
        return fmt.Errorf("failed to get public key: %w", err)
    }

    // Encrypt the secret value
    encryptedValue, err := g.encryptor.Encrypt(value, pubKey)
    if err != nil {
        return fmt.Errorf("failed to encrypt secret: %w", err)
    }

    // Create the encrypted secret
    secret := &github.EncryptedSecret{
        Name:           name,
        KeyID:         *pubKey.KeyID,
        EncryptedValue: encryptedValue,
    }

    // Add the secret to the repository
    _, err = g.client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
    if err != nil {
        return fmt.Errorf("failed to create secret: %w", err)
    }

    g.logger.Info("Secret added successfully", "repo", fmt.Sprintf("%s/%s", owner, repo), "secret", name)
    return nil
}

// Workflow Operations

// AddWorkflow adds a workflow file to a repository
func (g *GitHubOperations) AddWorkflow(ctx context.Context, owner, repo, path, content string) error {
    // Ensure path is in .github/workflows
    if !strings.HasPrefix(path, ".github/workflows/") {
        path = fmt.Sprintf(".github/workflows/%s", path)
    }
    if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
        path = fmt.Sprintf("%s.yml", path)
    }

    // Create commit message
    message := fmt.Sprintf("Add workflow: %s", path)

    // Create or update file
    opts := &github.RepositoryContentFileOptions{
        Message: &message,
        Content: []byte(content),
    }

    // Check if file exists first
    _, _, err := g.client.Repositories.GetContents(ctx, owner, repo, path, nil)
    if err == nil {
        // File exists, update it
        _, _, err = g.client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
    } else {
        // File doesn't exist, create it
        _, _, err = g.client.Repositories.CreateFile(ctx, owner, repo, path, opts)
    }

    if err != nil {
        return fmt.Errorf("failed to create workflow file: %w", err)
    }

    g.logger.Info("Workflow added successfully", "repo", fmt.Sprintf("%s/%s", owner, repo), "path", path)
    return nil
}

// TriggerWorkflow triggers a workflow run
func (g *GitHubOperations) TriggerWorkflow(ctx context.Context, owner, repo, workflowID string, ref string, inputs map[string]interface{}) (*github.WorkflowRun, error) {
    // Create workflow dispatch event
    event := github.CreateWorkflowDispatchEventRequest{
        Ref:    ref,
        Inputs: inputs,
    }

    // Trigger the workflow
    err := g.client.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowID, event)
    if err != nil {
        return nil, fmt.Errorf("failed to trigger workflow: %w", err)
    }

    // Get the triggered run (latest run for the workflow)
    runs, _, err := g.client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowID, &github.ListWorkflowRunsOptions{
        ListOptions: github.ListOptions{
            PerPage: 1,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get workflow run: %w", err)
    }

    if len(runs.WorkflowRuns) == 0 {
        return nil, fmt.Errorf("no workflow runs found")
    }

    g.logger.Info("Workflow triggered successfully", 
        "repo", fmt.Sprintf("%s/%s", owner, repo),
        "workflow", workflowID,
        "run_id", runs.WorkflowRuns[0].GetID())

    return runs.WorkflowRuns[0], nil
}

// File Operations

// CheckFile checks if a file exists in a repository
func (g *GitHubOperations) CheckFile(ctx context.Context, owner, repo, path string) (bool, *github.RepositoryContent, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("file:%s/%s/%s", owner, repo, path)
    if cached, ok := g.cache.Get(cacheKey); ok {
        if content, ok := cached.(*github.RepositoryContent); ok {
            return true, content, nil
        }
    }

    // Get file content from GitHub
    content, _, _, err := g.client.Repositories.GetContents(ctx, owner, repo, path, nil)
    if err != nil {
        if strings.Contains(err.Error(), "404") {
            return false, nil, nil
        }
        return false, nil, fmt.Errorf("failed to check file: %w", err)
    }

    // Cache the result
    g.cache.Set(cacheKey, content, defaultCacheDuration)

    return true, content, nil
}

// GetFileContent gets the decoded content of a file
func (g *GitHubOperations) GetFileContent(ctx context.Context, owner, repo, path string) (string, error) {
    exists, content, err := g.CheckFile(ctx, owner, repo, path)
    if err != nil {
        return "", err
    }
    if !exists {
        return "", fmt.Errorf("file not found: %s", path)
    }

    // Get the decoded content
    decoded, err := base64.StdEncoding.DecodeString(*content.Content)
    if err != nil {
        return "", fmt.Errorf("failed to decode content: %w", err)
    }

    return string(decoded), nil
}

// Usage Examples:

// Example workflow using the operations
func ExampleUsage() {
    ctx := context.Background()
    ghOps := NewGitHubOperations(os.Getenv("GITHUB_TOKEN"), NewLogger(), NewCache())

    // Add a secret
    err := ghOps.AddSecret(ctx, "owner", "repo", "API_KEY", "secret-value")
    if err != nil {
        log.Fatal(err)
    }

    // Add a workflow
    workflowContent := `
name: CI
on:
  push:
    branches: [ main ]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    `
    err = ghOps.AddWorkflow(ctx, "owner", "repo", "ci.yml", workflowContent)
    if err != nil {
        log.Fatal(err)
    }

    // Trigger the workflow
    inputs := map[string]interface{}{
        "environment": "production",
    }
    run, err := ghOps.TriggerWorkflow(ctx, "owner", "repo", "ci.yml", "main", inputs)
    if err != nil {
        log.Fatal(err)
    }

    // Check for .nodeprop.yml
    exists, _, err := ghOps.CheckFile(ctx, "owner", "repo", ".nodeprop.yml")
    if err != nil {
        log.Fatal(err)
    }
    if !exists {
        log.Println(".nodeprop.yml not found")
    }
}