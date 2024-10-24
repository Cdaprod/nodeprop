// pkg/nodeprop/events.go
package nodeprop

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// EventType represents different types of events
type EventType string

const (
    EventTypeNodeProp  EventType = "nodeprop"
    EventTypeWorkflow  EventType = "workflow"
    EventTypeSecret    EventType = "secret"
    EventTypeConfig    EventType = "config"
    EventTypeError     EventType = "error"
    EventTypeSystem    EventType = "system"
)

// Event represents a system event
type Event struct {
    ID        string                 `json:"id"`
    Type      EventType             `json:"type"`
    Name      string                `json:"name"`
    Data      interface{}           `json:"data"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time             `json:"timestamp"`
}

// EventHandler represents a function that handles events
type EventHandler func(Event) error

// EventBus manages event publishing and subscription
type EventBus struct {
    subscribers map[EventType]map[string]EventHandler
    middleware  []EventMiddleware
    consumer    EventConsumer
    mu          sync.RWMutex
    logger      Logger
}

// EventMiddleware represents a function that processes events before delivery
type EventMiddleware func(Event) Event

// EventSubscription represents an active subscription
type EventSubscription struct {
    ID       string
    Type     EventType
    Handler  EventHandler
    unsubFn  func()
}

// NewEventBus creates a new event bus instance
func NewEventBus(logger Logger) *EventBus {
    return &EventBus{
        subscribers: make(map[EventType]map[string]EventHandler),
        middleware:  make([]EventMiddleware, 0),
        logger:     logger,
    }
}

// Subscribe registers a handler for specific event types
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) *EventSubscription {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    if eb.subscribers[eventType] == nil {
        eb.subscribers[eventType] = make(map[string]EventHandler)
    }

    id := fmt.Sprintf("%s-%s", eventType, uuid.New().String())
    eb.subscribers[eventType][id] = handler

    return &EventSubscription{
        ID:      id,
        Type:    eventType,
        Handler: handler,
        unsubFn: func() {
            eb.unsubscribe(eventType, id)
        },
    }
}

// Unsubscribe removes a subscription
func (sub *EventSubscription) Unsubscribe() {
    sub.unsubFn()
}

func (eb *EventBus) unsubscribe(eventType EventType, id string) {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    if handlers, ok := eb.subscribers[eventType]; ok {
        delete(handlers, id)
    }
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(ctx context.Context, event Event) {
    // Ensure timestamp is set
    if event.Timestamp.IsZero() {
        event.Timestamp = time.Now()
    }

    // Apply middleware
    for _, mw := range eb.middleware {
        event = mw(event)
    }

    eb.mu.RLock()
    handlers := eb.subscribers[event.Type]
    eb.mu.RUnlock()

    // Fan out to all subscribers
    for id, handler := range handlers {
        go func(id string, h EventHandler) {
            if err := h(event); err != nil {
                eb.logger.WithField("subscriber", id).
                    WithField("event_type", event.Type).
                    WithField("event_name", event.Name).
                    Error("Failed to handle event:", err)
            }
        }(id, handler)
    }
}

// AddMiddleware adds event processing middleware
func (eb *EventBus) AddMiddleware(mw EventMiddleware) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.middleware = append(eb.middleware, mw)
}

// EventStream provides a channel of events
type EventStream struct {
    events chan Event
    done   chan struct{}
}

// NewEventStream creates a new event stream
func (eb *EventBus) NewEventStream(ctx context.Context, types ...EventType) *EventStream {
    stream := &EventStream{
        events: make(chan Event, 100),
        done:   make(chan struct{}),
    }

    // Subscribe to all requested event types
    for _, t := range types {
        sub := eb.Subscribe(t, func(e Event) error {
            select {
            case stream.events <- e:
                return nil
            case <-ctx.Done():
                return ctx.Err()
            case <-stream.done:
                return fmt.Errorf("stream closed")
            }
        })

        // Cleanup subscription when context is done
        go func() {
            select {
            case <-ctx.Done():
                sub.Unsubscribe()
            case <-stream.done:
                sub.Unsubscribe()
            }
        }()
    }

    return stream
}

// Events returns the event channel
func (es *EventStream) Events() <-chan Event {
    return es.events
}

// Close closes the event stream
func (es *EventStream) Close() {
    close(es.done)
}

// Utility functions for common events
func NewErrorEvent(err error) Event {
    return Event{
        ID:        uuid.New().String(),
        Type:      EventTypeError,
        Name:      "Error",
        Data:      err,
        Timestamp: time.Now(),
    }
}

func NewNodePropEvent(name string, data interface{}) Event {
    return Event{
        ID:        uuid.New().String(),
        Type:      EventTypeNodeProp,
        Name:      name,
        Data:      data,
        Timestamp: time.Now(),
    }
}

// Example middleware
func LoggingMiddleware(logger Logger) EventMiddleware {
    return func(e Event) Event {
        logger.WithField("event_type", e.Type).
            WithField("event_name", e.Name).
            Debug("Event processed")
        return e
    }
}

func MetricsMiddleware(metrics MetricsCollector) EventMiddleware {
    return func(e Event) Event {
        metrics.IncrementCounter(fmt.Sprintf("events_%s_total", e.Type))
        return e
    }
}

// EventConsumer defines how events should be handled
type EventConsumer interface {
    Consume(context.Context, Event) error
}

// RegistryEventConsumer sends events to a registry service
type RegistryEventConsumer struct {
    client    RegistryClient
    logger    Logger
    batchSize int
    events    chan Event
}

// LocalEventConsumer logs events locally
type LocalEventConsumer struct {
    logger Logger
    store  Store
}

// MultiEventConsumer allows multiple consumers
type MultiEventConsumer struct {
    consumers []EventConsumer
}

func NewRegistryEventConsumer(client RegistryClient, logger Logger) *RegistryEventConsumer {
    return &RegistryEventConsumer{
        client:    client,
        logger:    logger,
        batchSize: 100,
        events:    make(chan Event, 1000),
    }
}

func (rec *RegistryEventConsumer) Start(ctx context.Context) {
    go rec.processEvents(ctx)
}

func (rec *RegistryEventConsumer) Consume(ctx context.Context, event Event) error {
    select {
    case rec.events <- event:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Channel full, log warning and drop event
        rec.logger.Warn("Event channel full, dropping event")
        return fmt.Errorf("event channel full")
    }
}

func (rec *RegistryEventConsumer) processEvents(ctx context.Context) {
    batch := make([]Event, 0, rec.batchSize)
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case event := <-rec.events:
            batch = append(batch, event)
            if len(batch) >= rec.batchSize {
                rec.sendBatch(ctx, batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                rec.sendBatch(ctx, batch)
                batch = batch[:0]
            }
        }
    }
}

func (rec *RegistryEventConsumer) sendBatch(ctx context.Context, batch []Event) {
    if err := rec.client.SendEvents(ctx, batch); err != nil {
        rec.logger.WithError(err).Error("Failed to send events to registry")
        // Store failed events for retry
        rec.storeFailedEvents(batch)
    }
}

func NewLocalEventConsumer(logger Logger, store Store) *LocalEventConsumer {
    return &LocalEventConsumer{
        logger: logger,
        store:  store,
    }
}

func (lec *LocalEventConsumer) Consume(ctx context.Context, event Event) error {
    // Log the event
    lec.logger.WithFields(map[string]interface{}{
        "event_type": event.Type,
        "event_name": event.Name,
        "timestamp": event.Timestamp,
    }).Info("Event received")

    // Store event if needed
    if shouldStore(event) {
        key := fmt.Sprintf("events:%s:%s", event.Type, event.ID)
        if err := lec.store.Set(key, event); err != nil {
            lec.logger.WithError(err).Error("Failed to store event")
        }
    }

    return nil
}

func NewEventBus(logger Logger, consumer EventConsumer) *EventBus {
    if consumer == nil {
        // Default to local consumer if none provided
        consumer = NewLocalEventConsumer(logger, NewFileStore())
    }

    return &EventBus{
        subscribers: make(map[EventType]map[string]EventHandler),
        middleware:  make([]EventMiddleware, 0),
        logger:     logger,
        consumer:   consumer,
    }
}

// Updated Publish method
func (eb *EventBus) Publish(ctx context.Context, event Event) {
    // Process middleware
    for _, mw := range eb.middleware {
        event = mw(event)
    }

    // Send to consumer
    if err := eb.consumer.Consume(ctx, event); err != nil {
        eb.logger.WithError(err).Error("Failed to consume event")
    }

    // Notify subscribers
    eb.mu.RLock()
    handlers := eb.subscribers[event.Type]
    eb.mu.RUnlock()

    for id, handler := range handlers {
        go func(id string, h EventHandler) {
            if err := h(event); err != nil {
                eb.logger.WithField("subscriber", id).
                    WithError(err).
                    Error("Failed to handle event")
            }
        }(id, handler)
    }
}

// Helper functions for event handling
func shouldStore(event Event) bool {
    switch event.Type {
    case EventTypeNodeProp, EventTypeWorkflow, EventTypeSecret:
        return true
    default:
        return false
    }
}

// Example usage in manager.go
func NewNodePropManager(ctx context.Context, opts ...Option) (*NodePropManager, error) {
    m := &NodePropManager{
        config: DefaultConfig(),
        logger: NewLogger(),
    }

    // Apply options
    for _, opt := range opts {
        if err := opt(m); err != nil {
            return nil, err
        }
    }

    // Initialize event system based on context
    var consumer EventConsumer
    if registryClient := GetRegistryClientFromContext(ctx); registryClient != nil {
        consumer = NewRegistryEventConsumer(registryClient, m.logger)
    } else {
        consumer = NewLocalEventConsumer(m.logger, NewFileStore())
    }

    m.eventBus = NewEventBus(m.logger, consumer)

    return m, nil
}

// Example registry client interface
type RegistryClient interface {
    SendEvents(ctx context.Context, events []Event) error
}

// Context utilities
func WithRegistryClient(ctx context.Context, client RegistryClient) context.Context {
    return context.WithValue(ctx, registryClientKey, client)
}

func GetRegistryClientFromContext(ctx context.Context) RegistryClient {
    if client, ok := ctx.Value(registryClientKey).(RegistryClient); ok {
        return client
    }
    return nil
}