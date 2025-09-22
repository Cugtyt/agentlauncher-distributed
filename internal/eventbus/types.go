package eventbus

import "context"

type Event interface {
	Subject() string
}

type EventHandler[T Event] func(context.Context, T)

type EventBus interface {
	Emit(event Event) error
	Close() error
}
