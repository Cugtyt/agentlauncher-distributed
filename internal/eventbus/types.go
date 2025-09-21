package eventbus

import "context"

type Event any

type EventHandler func(context.Context, []byte)

type EventBus interface {
	Emit(subject string, event Event) error
	Subscribe(subject, queue string, handler EventHandler) error
	Close() error
}
