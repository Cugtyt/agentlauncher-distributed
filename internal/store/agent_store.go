package store

import (
	"encoding/json"
	"fmt"

	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type AgentData struct {
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	SystemPrompt string                    `json:"system_prompt"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	Messages     llminterface.MessageList  `json:"messages"`
}

type AgentStore struct {
	redis *RedisClient
}

func (as *AgentStore) agentKey(agentID string) string {
	return fmt.Sprintf("agent:%s", agentID)
}

func NewAgentStore(redisURL string) (*AgentStore, error) {
	redisClient, err := NewRedisClient(redisURL)
	if err != nil {
		return nil, err
	}
	return &AgentStore{
		redis: redisClient,
	}, nil
}

func (as *AgentStore) SetAgentData(agentID string, agentData *AgentData) error {
	data, err := json.Marshal(agentData)
	if err != nil {
		return fmt.Errorf("failed to marshal agent data: %w", err)
	}

	if err := as.redis.HSet(as.agentKey(agentID), "data", string(data)); err != nil {
		return fmt.Errorf("failed to store agent data: %w", err)
	}

	return nil
}

func (as *AgentStore) GetAgentData(agentID string) (*AgentData, error) {
	data, err := as.redis.HGet(as.agentKey(agentID), "data")
	if err != nil {
		return nil, fmt.Errorf("failed to get agent data: %w", err)
	}

	var agent AgentData
	if err := json.Unmarshal([]byte(data), &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent data: %w", err)
	}

	return &agent, nil
}

func (as *AgentStore) SetConversation(agentID string, messages llminterface.MessageList) error {
	messagesData, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal messages: %w", err)
	}

	if err := as.redis.HSet(as.agentKey(agentID), "messages", string(messagesData)); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

func (as *AgentStore) GetConversation(agentID string) (llminterface.MessageList, error) {
	messagesStr, err := as.redis.HGet(as.agentKey(agentID), "messages")
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	var messages llminterface.MessageList
	if err := json.Unmarshal([]byte(messagesStr), &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal messages: %w", err)
	}

	return messages, nil
}

func (as *AgentStore) Exists(agentID string) (bool, error) {
	exists, err := as.redis.Exists(as.agentKey(agentID))
	if err != nil {
		return false, fmt.Errorf("failed to check if agent exists: %w", err)
	}
	return exists > 0, nil
}

func (as *AgentStore) Delete(agentID string) error {
	if err := as.redis.Del(as.agentKey(agentID)); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

func (as *AgentStore) CreateAgent(agentData *AgentData) error {
	return as.SetAgentData(agentData.AgentID, agentData)
}

func (as *AgentStore) GetAgent(agentID string) (*AgentData, error) {
	return as.GetAgentData(agentID)
}

func (as *AgentStore) Close() error {
	return as.redis.Close()
}
