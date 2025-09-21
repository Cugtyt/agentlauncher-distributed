package eventbus

import (
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"
)

type DistributedEventBus struct {
    nats         *nats.Conn
    jetStream    nats.JetStreamContext
    subscriptions []*nats.Subscription
}

func NewDistributedEventBus(natsURL string) (*DistributedEventBus, error) {
    // Connect to NATS with reconnection options
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

    // Initialize JetStream
    js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
    if err != nil {
        nc.Close()
        return nil, fmt.Errorf("failed to initialize JetStream: %w", err)
    }

    // Create the distributed event bus
    deb := &DistributedEventBus{
        nats:      nc,
        jetStream: js,
    }

    // Initialize streams
    if err := deb.initializeStreams(); err != nil {
        nc.Close()
        return nil, fmt.Errorf("failed to initialize streams: %w", err)
    }

    log.Printf("Connected to NATS at %s", natsURL)
    return deb, nil
}

func (deb *DistributedEventBus) initializeStreams() error {
    streams := []nats.StreamConfig{
        {
            Name:        "TASKS",
            Subjects:    []string{"task.>"},
            Retention:   nats.WorkQueuePolicy,
            Storage:     nats.FileStorage,
            Duplicates:  2 * time.Minute,
            MaxAge:      24 * time.Hour,
        },
        {
            Name:        "AGENTS",
            Subjects:    []string{"agent.>"},
            Retention:   nats.WorkQueuePolicy,
            Storage:     nats.FileStorage,
            Duplicates:  2 * time.Minute,
            MaxAge:      24 * time.Hour,
        },
        {
            Name:        "MESSAGES",
            Subjects:    []string{"message.>"},
            Retention:   nats.WorkQueuePolicy,
            Storage:     nats.FileStorage,
            Duplicates:  2 * time.Minute,
            MaxAge:      24 * time.Hour,
        },
        {
            Name:        "LLM",
            Subjects:    []string{"llm.>"},
            Retention:   nats.WorkQueuePolicy,
            Storage:     nats.FileStorage,
            Duplicates:  2 * time.Minute,
            MaxAge:      1 * time.Hour,
        },
        {
            Name:        "TOOLS",
            Subjects:    []string{"tool.>"},
            Retention:   nats.WorkQueuePolicy,
            Storage:     nats.FileStorage,
            Duplicates:  2 * time.Minute,
            MaxAge:      1 * time.Hour,
        },
    }

    for _, streamConfig := range streams {
        // Check if stream exists
        _, err := deb.jetStream.StreamInfo(streamConfig.Name)
        if err != nil {
            // Stream doesn't exist, create it
            _, err = deb.jetStream.AddStream(&streamConfig)
            if err != nil {
                return fmt.Errorf("failed to create stream %s: %w", streamConfig.Name, err)
            }
            log.Printf("Created JetStream stream: %s", streamConfig.Name)
        } else {
            log.Printf("JetStream stream %s already exists", streamConfig.Name)
        }
    }

    return nil
}

func (deb *DistributedEventBus) Emit(subject string, event Event) error {
    // Marshal event to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    // Publish to JetStream for guaranteed delivery
    _, err = deb.jetStream.Publish(subject, data)
    if err != nil {
        return fmt.Errorf("failed to publish event to %s: %w", subject, err)
    }

    log.Printf("EventBus: Event emitted to %s", subject)
    return nil
}

func (deb *DistributedEventBus) Subscribe(subject, queue string, handler EventHandler) error {
    // Create durable consumer with queue group for load balancing
    consumerName := fmt.Sprintf("%s-consumer", queue)
    
    sub, err := deb.jetStream.QueueSubscribe(subject, queue,
        func(msg *nats.Msg) {
            // Handle the event
            handler(nil, msg.Data)
            
            // Acknowledge the message
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

    // Store subscription for cleanup
    deb.subscriptions = append(deb.subscriptions, sub)
    
    log.Printf("EventBus: Subscribed to %s with queue %s", subject, queue)
    return nil
}

func (deb *DistributedEventBus) Close() error {
    log.Println("Closing EventBus connections...")
    
    // Unsubscribe from all subscriptions
    for _, sub := range deb.subscriptions {
        if err := sub.Unsubscribe(); err != nil {
            log.Printf("Error unsubscribing: %v", err)
        }
    }
    
    // Close NATS connection
    if deb.nats != nil {
        deb.nats.Close()
    }
    
    log.Println("EventBus closed")
    return nil
}

// Utility method to check if connected
func (deb *DistributedEventBus) IsConnected() bool {
    return deb.nats != nil && deb.nats.IsConnected()
}

// Utility method to get connection status
func (deb *DistributedEventBus) Status() string {
    if deb.nats == nil {
        return "Not initialized"
    }
    if deb.nats.IsConnected() {
        return fmt.Sprintf("Connected to %s", deb.nats.ConnectedUrl())
    }
    return "Disconnected"
}