// cmd/tui/state/store.go
package state

import (
    "context"
    "sync"
    "time"

    "github.com/Cdaprod/nodeprop/pkg/nodeprop"
)

// Store manages the global state for the TUI
type Store struct {
    mu      sync.RWMutex
    state   State
    manager *nodeprop.NodePropManager
    subs    map[string][]chan State
}

// State represents the global TUI state
type State struct {
    Workflows     []nodeprop.Workflow
    Secrets       []nodeprop.Secret
    Files         []string
    Config        map[string]interface{}
    LoadingStates map[string]bool
    Errors        map[string]error
    CurrentRepo   string
    LastUpdated   time.Time
}

// Action represents state modifications
type Action interface {
    Apply(*State)
}

func NewStore(manager *nodeprop.NodePropManager) *Store {
    return &Store{
        manager: manager,
        subs:    make(map[string][]chan State),
        state: State{
            LoadingStates: make(map[string]bool),
            Errors:       make(map[string]error),
        },
    }
}

// Subscribe to state changes
func (s *Store) Subscribe(id string) chan State {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    ch := make(chan State, 1)
    s.subs[id] = append(s.subs[id], ch)
    
    // Send current state immediately
    ch <- s.state
    
    return ch
}

// Unsubscribe from state changes
func (s *Store) Unsubscribe(id string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    for _, ch := range s.subs[id] {
        close(ch)
    }
    delete(s.subs, id)
}

// Dispatch applies an action and notifies subscribers
func (s *Store) Dispatch(action Action) {
    s.mu.Lock()
    action.Apply(&s.state)
    s.state.LastUpdated = time.Now()
    s.mu.Unlock()

    // Notify subscribers
    s.notify()
}

func (s *Store) notify() {
    s.mu.RLock()
    state := s.state
    s.mu.RUnlock()

    for _, channels := range s.subs {
        for _, ch := range channels {
            select {
            case ch <- state:
            default:
                // Skip if channel is blocked
            }
        }
    }
}

// Actions
type LoadWorkflowsAction struct {
    Workflows []nodeprop.Workflow
}

func (a LoadWorkflowsAction) Apply(s *State) {
    s.Workflows = a.Workflows
    s.LoadingStates["workflows"] = false
}

type LoadSecretsAction struct {
    Secrets []nodeprop.Secret
}

func (a LoadSecretsAction) Apply(s *State) {
    s.Secrets = a.Secrets
    s.LoadingStates["secrets"] = false
}

// Async operations
func (s *Store) LoadWorkflows(ctx context.Context, repo string) error {
    s.Dispatch(SetLoadingAction{"workflows", true})
    
    workflows, err := s.manager.ListWorkflows(ctx, repo)
    if err != nil {
        s.Dispatch(SetErrorAction{"workflows", err})
        return err
    }
    
    s.Dispatch(LoadWorkflowsAction{workflows})
    return nil
}