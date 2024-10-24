// pkg/nodeprop/validator.go
package nodeprop

import (
    "fmt"
    "regexp"
)

// ValidationRule defines a single validation rule
type ValidationRule struct {
    Field     string
    Validator func(interface{}) error
    Message   string
}

// NodePropValidator handles validation of NodeProp configurations
type NodePropValidator struct {
    rules map[string][]ValidationRule
}

func NewNodePropValidator() *NodePropValidator {
    v := &NodePropValidator{
        rules: make(map[string][]ValidationRule),
    }
    
    // Add default validation rules
    v.AddRule("repository", ValidationRule{
        Field: "Name",
        Validator: func(i interface{}) error {
            name, ok := i.(string)
            if !ok || name == "" {
                return fmt.Errorf("invalid repository name")
            }
            matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", name)
            if !matched {
                return fmt.Errorf("repository name contains invalid characters")
            }
            return nil
        },
        Message: "Repository name must contain only alphanumeric characters, hyphens, and underscores",
    })

    return v
}