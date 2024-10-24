// cmd/nodeprop/tui/views/workflows.go
type WorkflowsView struct {
    manager    *nodeprop.NodePropManager
    workflows  []nodeprop.Workflow
    selected   int
    width     int
    height    int
    loading   bool
    err       error
}

func NewWorkflowsView(manager *nodeprop.NodePropManager) *WorkflowsView {
    return &WorkflowsView{
        manager: manager,
    }
}

func (v *WorkflowsView) Init() tea.Cmd {
    return v.loadWorkflows
}

func (v *WorkflowsView) loadWorkflows() tea.Msg {
    ctx := context.Background()
    
    // Set loading state
    v.loading = true
    
    go func() {
        workflows, err := v.manager.ListWorkflows(ctx, v.currentRepo)
        if err != nil {
            v.program.Send(errMsg{err})
            return
        }
        
        v.program.Send(workflowsLoadedMsg{workflows})
    }()
    
    return nil
}

// Message types for async operations
type workflowsLoadedMsg struct {
    workflows []nodeprop.Workflow
}

// Update the Update method to handle the messages
func (v *WorkflowsView) Update(msg tea.Msg) (View, tea.Cmd) {
    switch msg := msg.(type) {
    case workflowsLoadedMsg:
        v.workflows = msg.workflows
        v.loading = false
        return v, nil
        
    case errMsg:
        v.err = msg.err
        v.loading = false
        return v, nil
        
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, v.keys.Enter):
            if v.selected >= 0 && v.selected < len(v.workflows) {
                return v, v.triggerWorkflow(v.workflows[v.selected])
            }
        }
    }
    
    return v, nil
}

// Add workflow triggering
func (v *WorkflowsView) triggerWorkflow(workflow nodeprop.Workflow) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        err := v.manager.TriggerWorkflow(ctx, workflow.ID)
        if err != nil {
            return errMsg{err}
        }
        return workflowTriggeredMsg{workflow.ID}
    }
}

// Update the View method to show loading state and errors
func (v *WorkflowsView) View() string {
    var s strings.Builder
    
    s.WriteString(titleStyle.Render("Workflows"))
    s.WriteString("\n\n")
    
    if v.loading {
        s.WriteString("Loading workflows...")
        return s.String()
    }
    
    if v.err != nil {
        s.WriteString(errorStyle.Render(v.err.Error()))
        return s.String()
    }
    
    for i, workflow := range v.workflows {
        style := itemStyle
        if i == v.selected {
            style = selectedItemStyle
        }
        
        s.WriteString(style.Render(fmt.Sprintf(
            "%s (%s)\n",
            workflow.Name,
            workflow.Status,
        )))
    }
    
    return s.String()
}