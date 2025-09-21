package eventbus

import "context"

type Event interface {
	Subject() string
}

type EventHandler func(context.Context, []byte)

type EventBus interface {
	Emit(event Event) error
	Subscribe(subject, queue string, handler EventHandler) error
	Close() error
}
