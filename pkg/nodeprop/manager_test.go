// pkg/nodeprop/manager_test.go
package nodeprop

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

// MockStore implements Store interface for testing
type MockStore struct {
    mock.Mock
}

func (m *MockStore) Get(key string) (interface{}, error) {
    args := m.Called(key)
    return args.Get(0), args.Error(1)
}

func (m *MockStore) Set(key string, value interface{}) error {
    args := m.Called(key, value)
    return args.Error(0)
}

func (m *MockStore) Delete(key string) error {
    args := m.Called(key)
    return args.Error(0)
}

func (m *MockStore) List(prefix string) (map[string]interface{}, error) {
    args := m.Called(prefix)
    return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockGitHubClient implements GitHub operations for testing
type MockGitHubClient struct {
    mock.Mock
}

func (m *MockGitHubClient) CreateWorkflow(ctx context.Context, args WorkflowArguments) error {
    mockArgs := m.Called(ctx, args)
    return mockArgs.Error(0)
}

func (m *MockGitHubClient) ListWorkflows(ctx context.Context, repo string) ([]Workflow, error) {
    args := m.Called(ctx, repo)
    return args.Get(0).([]Workflow), args.Error(1)
}

// Helper function to create a test environment
func setupTest(t *testing.T) (*NodePropManager, *MockStore, *MockGitHubClient, string, func()) {
    tempDir, err := ioutil.TempDir("", "nodeprop-test-*")
    require.NoError(t, err)

    mockStore := new(MockStore)
    mockGitHub := new(MockGitHubClient)

    manager, err := NewNodePropManager(
        context.Background(),
        WithStore(mockStore),
        WithGitHubToken("test-token"),
    )
    require.NoError(t, err)

    // Replace the GitHub client with our mock
    manager.githubClient = mockGitHub

    cleanup := func() {
        os.RemoveAll(tempDir)
    }

    return manager, mockStore, mockGitHub, tempDir, cleanup
}

func TestAddWorkflow(t *testing.T) {
    manager, _, mockGitHub, tempDir, cleanup := setupTest(t)
    defer cleanup()

    ctx := context.Background()
    args := WorkflowArguments{
        Repository: "owner/repo",
        Name:      "test-workflow.yml",
        Content:   "name: Test\non: push",
        Template:  "",
        Variables: nil,
    }

    // Test successful workflow addition
    t.Run("successful workflow addition", func(t *testing.T) {
        mockGitHub.On("CreateWorkflow", ctx, args).Return(nil)

        err := manager.AddWorkflow(ctx, args)
        assert.NoError(t, err)
        mockGitHub.AssertExpectations(t)
    })

    // Test workflow addition with error
    t.Run("workflow addition error", func(t *testing.T) {
        mockErr := fmt.Errorf("github error")
        mockGitHub.On("CreateWorkflow", ctx, args).Return(mockErr)

        err := manager.AddWorkflow(ctx, args)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "github error")
    })
}

func TestGenerateNodeProp(t *testing.T) {
    manager, mockStore, _, tempDir, cleanup := setupTest(t)
    defer cleanup()

    ctx := context.Background()
    args := NodePropArguments{
        RepoPath: filepath.Join(tempDir, "test-repo"),
        RepoName: "test-repo",
        Domain:   "test.example.com",
    }

    // Test successful NodeProp generation
    t.Run("successful nodeprop generation", func(t *testing.T) {
        mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)

        err := manager.GenerateNodeProp(ctx, args)
        assert.NoError(t, err)

        // Verify the .nodeprop.yml was created
        nodePropPath := filepath.Join(args.RepoPath, ".nodeprop.yml")
        assert.FileExists(t, nodePropPath)

        // Verify content
        content, err := ioutil.ReadFile(nodePropPath)
        assert.NoError(t, err)
        assert.Contains(t, string(content), args.Domain)
        assert.Contains(t, string(content), args.RepoName)

        mockStore.AssertExpectations(t)
    })

    // Test NodeProp generation with invalid path
    t.Run("invalid path error", func(t *testing.T) {
        invalidArgs := args
        invalidArgs.RepoPath = "/nonexistent/path"

        err := manager.GenerateNodeProp(ctx, invalidArgs)
        assert.Error(t, err)
    })
}

func TestAddSecret(t *testing.T) {
    manager, mockStore, mockGitHub, _, cleanup := setupTest(t)
    defer cleanup()

    ctx := context.Background()
    args := SecretArguments{
        Repository: "owner/repo",
        Name:      "TEST_SECRET",
        Value:     "secret-value",
        Visibility: "private",
    }

    // Test successful secret addition
    t.Run("successful secret addition", func(t *testing.T) {
        mockStore.On("Set", mock.Anything, mock.Anything).Return(nil)
        mockGitHub.On("CreateSecret", ctx, args).Return(nil)

        err := manager.AddSecret(ctx, args)
        assert.NoError(t, err)

        mockStore.AssertExpectations(t)
        mockGitHub.AssertExpectations(t)
    })

    // Test secret addition with error
    t.Run("secret addition error", func(t *testing.T) {
        mockErr := fmt.Errorf("github error")
        mockGitHub.On("CreateSecret", ctx, args).Return(mockErr)

        err := manager.AddSecret(ctx, args)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "github error")
    })
}

func TestEvents(t *testing.T) {
    manager, _, _, _, cleanup := setupTest(t)
    defer cleanup()

    // Test event subscription and emission
    t.Run("event subscription and emission", func(t *testing.T) {
        events, unsub := manager.Subscribe(EventWorkflow)
        defer unsub()

        // Create test event
        testEvent := Event{
            Type:      EventWorkflow,
            Name:      "test-event",
            Data:      "test-data",
            Timestamp: time.Now(),
        }

        // Emit event in goroutine to prevent blocking
        go func() {
            manager.Emit(testEvent)
        }()

        // Wait for event
        select {
        case receivedEvent := <-events:
            assert.Equal(t, testEvent.Type, receivedEvent.Type)
            assert.Equal(t, testEvent.Name, receivedEvent.Name)
            assert.Equal(t, testEvent.Data, receivedEvent.Data)
        case <-time.After(time.Second):
            t.Fatal("timeout waiting for event")
        }
    })
}

func TestConfigManagement(t *testing.T) {
    manager, mockStore, _, _, cleanup := setupTest(t)
    defer cleanup()

    // Test configuration loading
    t.Run("load configuration", func(t *testing.T) {
        mockConfig := map[string]interface{}{
            "github.token": "test-token",
            "cache.ttl":   "1h",
        }
        mockStore.On("Get", "config").Return(mockConfig, nil)

        err := manager.LoadConfig(context.Background())
        assert.NoError(t, err)
        assert.Equal(t, "test-token", manager.GetConfigValue("github.token"))
    })

    // Test configuration saving
    t.Run("save configuration", func(t *testing.T) {
        mockStore.On("Set", "config", mock.Anything).Return(nil)

        manager.SetConfigValue("test.key", "test-value")
        err := manager.SaveConfig(context.Background())
        assert.NoError(t, err)
        mockStore.AssertExpectations(t)
    })
}

func TestValidation(t *testing.T) {
    manager, _, _, _, cleanup := setupTest(t)
    defer cleanup()

    // Test NodeProp validation
    t.Run("validate nodeprop configuration", func(t *testing.T) {
        validConfig := NodePropFile{
            ID:      uuid.New().String(),
            Name:    "test-repo",
            Address: "https://github.com/owner/test-repo",
            Status:  "active",
        }

        err := manager.ValidateNodeProp(context.Background(), validConfig)
        assert.NoError(t, err)

        invalidConfig := NodePropFile{
            // Missing required fields
        }

        err = manager.ValidateNodeProp(context.Background(), invalidConfig)
        assert.Error(t, err)
    })
}

func TestConcurrency(t *testing.T) {
    manager, _, _, _, cleanup := setupTest(t)
    defer cleanup()

    // Test concurrent operations
    t.Run("concurrent operations", func(t *testing.T) {
        ctx := context.Background()
        numGoroutines := 10
        done := make(chan bool)

        for i := 0; i < numGoroutines; i++ {
            go func(i int) {
                args := WorkflowArguments{
                    Repository: fmt.Sprintf("owner/repo-%d", i),
                    Name:      fmt.Sprintf("workflow-%d.yml", i),
                }
                _ = manager.AddWorkflow(ctx, args)
                done <- true
            }(i)
        }

        // Wait for all goroutines
        for i := 0; i < numGoroutines; i++ {
            <-done
        }
    })
}