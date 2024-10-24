// cmd/nodeprop/tui/views/files.go
type FilesView struct {
    manager  *nodeprop.NodePropManager
    files    []string
    selected int
    width    int
    height   int
    loading  bool
    err      error
}