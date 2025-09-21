package store

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/yourusername/agentlauncher-distributed/internal/events"
)

type MessageStore struct {
    redis *RedisClient
}

func NewMessageStore(redisURL string) *MessageStore {
    return &MessageStore{
        redis: NewRedisClient(redisURL),
    }
}

func (ms *MessageStore) AddMessages(agentID string, messages []events.Message) error {
    key := fmt.Sprintf("messages:%s", agentID)

    // Add each message to the list
    for _, msg := range messages {
        data, err := json.Marshal(msg)
        if err != nil {
            return fmt.Errorf("failed to marshal message: %w", err)
        }

        if err := ms.redis.RPush(key, data); err != nil {
            return fmt.Errorf("failed to add message: %w", err)
        }
    }

    // Set expiration (messages live for 1 hour after last update)
    if err := ms.redis.Expire(key, 1*time.Hour); err != nil {
        return fmt.Errorf("failed to set expiration: %w", err)
    }

    // Keep only last 200 messages to prevent memory bloat
    if err := ms.redis.LTrim(key, -200, -1); err != nil {
        return fmt.Errorf("failed to trim messages: %w", err)
    }

    return nil
}

func (ms *MessageStore) GetMessages(agentID string) ([]events.Message, error) {
    key := fmt.Sprintf("messages:%s", agentID)
    
    messageStrings, err := ms.redis.LRange(key, 0, -1)
    if err != nil {
        return nil, fmt.Errorf("failed to get messages: %w", err)
    }

    messages := make([]events.Message, 0, len(messageStrings))
    for _, msgStr := range messageStrings {
        var msg events.Message
        if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
            return nil, fmt.Errorf("failed to unmarshal message: %w", err)
        }
        messages = append(messages, msg)
    }

    return messages, nil
}

func (ms *MessageStore) GetRecentMessages(agentID string, count int) ([]events.Message, error) {
    key := fmt.Sprintf("messages:%s", agentID)
    
    // Get last 'count' messages
    start := int64(-count)
    messageStrings, err := ms.redis.LRange(key, start, -1)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent messages: %w", err)
    }

    messages := make([]events.Message, 0, len(messageStrings))
    for _, msgStr := range messageStrings {
        var msg events.Message
        if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
            return nil, fmt.Errorf("failed to unmarshal message: %w", err)
        }
        messages = append(messages, msg)
    }

    return messages, nil
}

func (ms *MessageStore) DeleteMessages(agentID string) error {
    key := fmt.Sprintf("messages:%s", agentID)
    return ms.redis.Del(key)
}

func (ms *MessageStore) GetMessageCount(agentID string) (int64, error) {
    key := fmt.Sprintf("messages:%s", agentID)
    return ms.redis.GetClient().LLen(ms.redis.GetContext(), key).Result()
}

func (ms *MessageStore) Close() error {
    return ms.redis.Close()
}