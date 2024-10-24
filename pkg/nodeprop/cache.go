// pkg/nodeprop/cache.go
type NodePropCache struct {
    cache    *lru.Cache
    Logger   *logrus.Logger
}

func NewNodePropCache(size int, logger *logrus.Logger) (*NodePropCache, error) {
    cache, err := lru.New(size)
    if err != nil {
        return nil, err
    }
    
    return &NodePropCache{
        cache:  cache,
        Logger: logger,
    }, nil
}
