package eventbus

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type DistributedEventBus struct {
	nats           *nats.Conn
	jetStream      nats.JetStreamContext
	subscriptions  []*nats.Subscription
	createdStreams map[string]bool
}

func NewDistributedEventBus(natsURL string) (*DistributedEventBus, error) {
	nc, err := nats.Connect(natsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %v", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to initialize JetStream: %w", err)
	}

	deb := &DistributedEventBus{
		nats:           nc,
		jetStream:      js,
		createdStreams: make(map[string]bool),
	}

	log.Printf("Connected to NATS at %s", natsURL)
	return deb, nil
}

func (deb *DistributedEventBus) ensureStreamForSubject(subject string) error {
	if deb.createdStreams[subject] {
		return nil
	}

	_, err := deb.jetStream.StreamInfo(subject)
	if err != nil {
		streamConfig := &nats.StreamConfig{
			Name:       subject,
			Subjects:   []string{subject},
			Retention:  nats.WorkQueuePolicy,
			Storage:    nats.FileStorage,
			Duplicates: 2 * time.Minute,
			MaxAge:     24 * time.Hour,
		}

		_, err = deb.jetStream.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("failed to create stream %s: %w", subject, err)
		}
		log.Printf("Created JetStream stream: %s", subject)
	}

	deb.createdStreams[subject] = true
	return nil
}

func (deb *DistributedEventBus) Emit(event Event) error {
	subject := event.Subject()

	if err := deb.ensureStreamForSubject(subject); err != nil {
		return fmt.Errorf("failed to ensure stream for %s: %w", subject, err)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = deb.jetStream.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event to %s: %w", subject, err)
	}

	log.Printf("EventBus: Event emitted to %s", subject)
	return nil
}

func (deb *DistributedEventBus) Subscribe(subject, queue string, handler EventHandler) error {
	if err := deb.ensureStreamForSubject(subject); err != nil {
		return err
	}

	consumerName := fmt.Sprintf("%s-consumer", queue)

	sub, err := deb.jetStream.QueueSubscribe(subject, queue,
		func(msg *nats.Msg) {
			handler(nil, msg.Data)
			msg.Ack()
		},
		nats.Durable(consumerName),
		nats.ManualAck(),
		nats.AckWait(30*time.Second),
		nats.MaxDeliver(3),
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	deb.subscriptions = append(deb.subscriptions, sub)

	log.Printf("EventBus: Subscribed to %s with queue %s", subject, queue)
	return nil
}

func (deb *DistributedEventBus) Close() error {
	log.Println("Closing EventBus connections...")

	for _, sub := range deb.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing: %v", err)
		}
	}

	if deb.nats != nil {
		deb.nats.Close()
	}

	log.Println("EventBus closed")
	return nil
}

func (deb *DistributedEventBus) IsConnected() bool {
	return deb.nats != nil && deb.nats.IsConnected()
}

func (deb *DistributedEventBus) Status() string {
	if deb.nats == nil {
		return "Not initialized"
	}
	if deb.nats.IsConnected() {
		return fmt.Sprintf("Connected to %s", deb.nats.ConnectedUrl())
	}
	return "Disconnected"
}
