// pkg/nodeprop/template.go
type TemplateManager struct {
    templates map[string]*template.Template
    Logger    *logrus.Logger
}

func NewTemplateManager(logger *logrus.Logger) *TemplateManager {
    return &TemplateManager{
        templates: make(map[string]*template.Template),
        Logger:    logger,
    }
}