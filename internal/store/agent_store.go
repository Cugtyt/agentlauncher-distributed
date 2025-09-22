package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type AgentData struct {
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	SystemPrompt string                    `json:"system_prompt"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
}

type AgentConversation struct {
	AgentID  string                   `json:"agent_id"`
	Messages llminterface.MessageList `json:"messages"`
}

type AgentStore struct {
	redis *RedisClient
}

func NewAgentStore(redisURL string) *AgentStore {
	return &AgentStore{
		redis: NewRedisClient(redisURL),
	}
}

func (as *AgentStore) SetAgentData(agentID string, agentData *AgentData) error {
	data, err := json.Marshal(agentData)
	if err != nil {
		return fmt.Errorf("failed to marshal agent data: %w", err)
	}

	agentKey := fmt.Sprintf("%s:data", agentID)
	if err := as.redis.Set(agentKey, data, 1*time.Hour); err != nil {
		return fmt.Errorf("failed to store agent data: %w", err)
	}

	return nil
}

func (as *AgentStore) GetAgentData(agentID string) (*AgentData, error) {
	agentKey := fmt.Sprintf("%s:data", agentID)
	data, err := as.redis.Get(agentKey)
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
	conversation := &AgentConversation{
		AgentID:  agentID,
		Messages: messages,
	}

	data, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	conversationKey := fmt.Sprintf("%s:conv", agentID)
	if err := as.redis.Set(conversationKey, data, 1*time.Hour); err != nil {
		return fmt.Errorf("failed to store conversation: %w", err)
	}

	return nil
}

func (as *AgentStore) GetConversation(agentID string) (llminterface.MessageList, error) {
	conversationKey := fmt.Sprintf("%s:conv", agentID)
	data, err := as.redis.Get(conversationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	var conversation AgentConversation
	if err := json.Unmarshal([]byte(data), &conversation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return conversation.Messages, nil
}

func (as *AgentStore) Exists(agentID string) (bool, error) {
	agentKey := fmt.Sprintf("%s:data", agentID)
	exists, err := as.redis.Exists(agentKey)
	if err != nil {
		return false, fmt.Errorf("failed to check if agent exists: %w", err)
	}
	return exists > 0, nil
}

func (as *AgentStore) Delete(agentID string) error {
	agentKey := fmt.Sprintf("%s:data", agentID)
	if err := as.redis.Del(agentKey); err != nil {
		return fmt.Errorf("failed to delete agent data: %w", err)
	}

	conversationKey := fmt.Sprintf("%s:conv", agentID)
	if err := as.redis.Del(conversationKey); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

func (as *AgentStore) CreateAgent(agentData *AgentData, conversation llminterface.MessageList) error {
	if err := as.SetAgentData(agentData.AgentID, agentData); err != nil {
		return err
	}

	if err := as.SetConversation(agentData.AgentID, conversation); err != nil {
		as.Delete(agentData.AgentID)
		return err
	}

	return nil
}

func (as *AgentStore) GetAgent(agentID string) (*AgentData, llminterface.MessageList, error) {
	agent, err := as.GetAgentData(agentID)
	if err != nil {
		return nil, nil, err
	}

	conversation, err := as.GetConversation(agentID)
	if err != nil {
		return nil, nil, err
	}

	return agent, conversation, nil
}

func (as *AgentStore) Close() error {
	return as.redis.Close()
}
