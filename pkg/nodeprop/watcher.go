// pkg/nodeprop/watcher.go
type ConfigWatcher struct {
    Logger *logrus.Logger
    changes chan ConfigChange
}

type ConfigChange struct {
    Type    string
    Path    string
    OldData interface{}
    NewData interface{}
}

func NewConfigWatcher(logger *logrus.Logger) *ConfigWatcher {
    return &ConfigWatcher{
        Logger:  logger,
        changes: make(chan ConfigChange, 100),
    }
}
