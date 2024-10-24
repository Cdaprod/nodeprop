// cmd/nodeprop/tui/views/secrets.go
type SecretsView struct {
    manager   *nodeprop.NodePropManager
    secrets   []nodeprop.Secret
    selected  int
    width    int
    height   int
    loading  bool
    err      error
}

