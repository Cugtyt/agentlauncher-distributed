package eventbus

import "context"

// Event is the base interface for all events
type Event interface{}

// EventHandler is a function that handles events
type EventHandler func(context.Context, []byte)

// DistributedEventBus interface
type DistributedEventBusInterface interface {
    Emit(subject string, event Event) error
    Subscribe(subject, queue string, handler EventHandler) error
    Close() error
}