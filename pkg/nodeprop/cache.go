```go
// pkg/nodeprop/cache.go
package nodeprop

import (
    "fmt"
    "sync"
    "time"
)

// CacheItem represents a single cached item
type CacheItem struct {
    Value      interface{}
    Expiration int64
}

// Cache implements thread-safe in-memory cache with TTL
type Cache struct {
    items map[string]CacheItem
    mu    sync.RWMutex
    
    // Default expiration for items
    defaultExpiration time.Duration
    
    // Cleanup interval for expired items
    cleanupInterval time.Duration
    
    // Optional size limit
    maxItems int
    
    // Statistics
    stats CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
    Hits      uint64
    Misses    uint64
    Evictions uint64
    Size      uint64
}

// CacheOption defines functional options for cache configuration
type CacheOption func(*Cache)

// NewCache creates a new cache instance with options
func NewCache(opts ...CacheOption) *Cache {
    c := &Cache{
        items:            make(map[string]CacheItem),
        defaultExpiration: 1 * time.Hour,
        cleanupInterval:   5 * time.Minute,
        maxItems:         10000, // Default max items
    }

    // Apply options
    for _, opt := range opts {
        opt(c)
    }

    // Start cleanup routine
    go c.startCleanup()

    return c
}

// WithExpiration sets default expiration time
func WithExpiration(d time.Duration) CacheOption {
    return func(c *Cache) {
        c.defaultExpiration = d
    }
}

// WithCleanupInterval sets cleanup interval
func WithCleanupInterval(d time.Duration) CacheOption {
    return func(c *Cache) {
        c.cleanupInterval = d
    }
}

// WithMaxItems sets maximum items limit
func WithMaxItems(n int) CacheOption {
    return func(c *Cache) {
        c.maxItems = n
    }
}

// Set adds an item to the cache
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
    var expiration int64

    if duration == 0 {
        duration = c.defaultExpiration
    }

    if duration > 0 {
        expiration = time.Now().Add(duration).UnixNano()
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // Check size limit before adding
    if len(c.items) >= c.maxItems {
        c.evictOldest()
    }

    c.items[key] = CacheItem{
        Value:      value,
        Expiration: expiration,
    }

    c.stats.Size = uint64(len(c.items))
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, found := c.items[key]
    if !found {
        c.stats.Misses++
        return nil, false
    }

    // Check if item has expired
    if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
        c.stats.Misses++
        return nil, false
    }

    c.stats.Hits++
    return item.Value, true
}

// GetWithExpiration returns the item and its expiration time
func (c *Cache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, found := c.items[key]
    if !found {
        c.stats.Misses++
        return nil, time.Time{}, false
    }

    if item.Expiration > 0 {
        if time.Now().UnixNano() > item.Expiration {
            c.stats.Misses++
            return nil, time.Time{}, false
        }
        return item.Value, time.Unix(0, item.Expiration), true
    }

    c.stats.Hits++
    return item.Value, time.Time{}, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    delete(c.items, key)
    c.stats.Size = uint64(len(c.items))
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.items = make(map[string]CacheItem)
    c.stats.Size = 0
}

// Items returns all unexpired items in the cache
func (c *Cache) Items() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()

    items := make(map[string]interface{}, len(c.items))
    now := time.Now().UnixNano()

    for k, v := range c.items {
        if v.Expiration == 0 || now < v.Expiration {
            items[k] = v.Value
        }
    }

    return items
}

// ItemCount returns the number of items in the cache
func (c *Cache) ItemCount() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return len(c.items)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.stats
}

// startCleanup starts the background cleanup routine
func (c *Cache) startCleanup() {
    ticker := time.NewTicker(c.cleanupInterval)
    defer ticker.Stop()

    for range ticker.C {
        c.DeleteExpired()
    }
}

// DeleteExpired removes expired items from the cache
func (c *Cache) DeleteExpired() {
    now := time.Now().UnixNano()
    c.mu.Lock()
    defer c.mu.Unlock()

    for k, v := range c.items {
        if v.Expiration > 0 && now > v.Expiration {
            delete(c.items, k)
            c.stats.Evictions++
        }
    }

    c.stats.Size = uint64(len(c.items))
}

// evictOldest removes the oldest item when cache is full
func (c *Cache) evictOldest() {
    var oldestKey string
    var oldestTime int64 = math.MaxInt64

    for k, v := range c.items {
        if v.Expiration < oldestTime {
            oldestKey = k
            oldestTime = v.Expiration
        }
    }

    if oldestKey != "" {
        delete(c.items, oldestKey)
        c.stats.Evictions++
    }
}

// Flush writes cache contents to persistent storage
func (c *Cache) Flush() error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // Example implementation - you might want to customize this
    for k, v := range c.items {
        if err := c.persistItem(k, v); err != nil {
            return fmt.Errorf("failed to persist cache item %s: %w", k, err)
        }
    }
    return nil
}

// persistItem writes a single cache item to storage
func (c *Cache) persistItem(key string, item CacheItem) error {
    // Implement persistence logic here
    return nil
}
```

