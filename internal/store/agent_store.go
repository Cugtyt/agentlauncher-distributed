package store

import (
    "encoding/json"
    "fmt"
    "time"
)

type AgentInfo struct {
    ID        string            `json:"id"`
    ParentID  string            `json:"parent_id,omitempty"`
    Status    string            `json:"status"` // "running", "completed", "failed"
    Task      string            `json:"task,omitempty"`
    Context   string            `json:"context,omitempty"`
    Tools     []string          `json:"tools,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
}

type AgentStore struct {
    redis *RedisClient
}

func NewAgentStore(redisURL string) *AgentStore {
    return &AgentStore{
        redis: NewRedisClient(redisURL),
    }
}

func (as *AgentStore) Create(agentID, parentID string) error {
    agent := &AgentInfo{
        ID:        agentID,
        ParentID:  parentID,
        Status:    "running",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    data, err := json.Marshal(agent)
    if err != nil {
        return fmt.Errorf("failed to marshal agent info: %w", err)
    }

    // Store agent info
    key := fmt.Sprintf("agent:%s", agentID)
    if err := as.redis.Set(key, data, 1*time.Hour); err != nil {
        return fmt.Errorf("failed to store agent: %w", err)
    }

    // Add to active agents set
    if err := as.redis.HSet("agents:active", agentID, "running"); err != nil {
        return fmt.Errorf("failed to add to active agents: %w", err)
    }

    // If this is a sub-agent, track parent relationship
    if parentID != "" {
        childrenKey := fmt.Sprintf("agent:%s:children", parentID)
        if err := as.redis.RPush(childrenKey, agentID); err != nil {
            return fmt.Errorf("failed to track parent relationship: %w", err)
        }
        as.redis.Expire(childrenKey, 1*time.Hour)
    }

    return nil
}

func (as *AgentStore) Get(agentID string) (*AgentInfo, error) {
    key := fmt.Sprintf("agent:%s", agentID)
    data, err := as.redis.Get(key)
    if err != nil {
        return nil, fmt.Errorf("failed to get agent: %w", err)
    }

    var agent AgentInfo
    if err := json.Unmarshal([]byte(data), &agent); err != nil {
        return nil, fmt.Errorf("failed to unmarshal agent info: %w", err)
    }

    return &agent, nil
}

func (as *AgentStore) Update(agentID string, updates map[string]interface{}) error {
    // Get current agent info
    agent, err := as.Get(agentID)
    if err != nil {
        return err
    }

    // Apply updates
    if status, ok := updates["status"].(string); ok {
        agent.Status = status
    }
    if task, ok := updates["task"].(string); ok {
        agent.Task = task
    }
    if context, ok := updates["context"].(string); ok {
        agent.Context = context
    }
    if metadata, ok := updates["metadata"].(map[string]string); ok {
        agent.Metadata = metadata
    }

    agent.UpdatedAt = time.Now()

    // Save updated agent
    data, err := json.Marshal(agent)
    if err != nil {
        return fmt.Errorf("failed to marshal updated agent: %w", err)
    }

    key := fmt.Sprintf("agent:%s", agentID)
    if err := as.redis.Set(key, data, 1*time.Hour); err != nil {
        return fmt.Errorf("failed to update agent: %w", err)
    }

    // Update active agents status
    if agent.Status != "running" {
        as.redis.HDel("agents:active", agentID)
    } else {
        as.redis.HSet("agents:active", agentID, agent.Status)
    }

    return nil
}

func (as *AgentStore) Delete(agentID string) error {
    // Remove from active agents
    if err := as.redis.HDel("agents:active", agentID); err != nil {
        return fmt.Errorf("failed to remove from active agents: %w", err)
    }

    // Delete agent data
    key := fmt.Sprintf("agent:%s", agentID)
    if err := as.redis.Del(key); err != nil {
        return fmt.Errorf("failed to delete agent: %w", err)
    }

    // Delete children tracking
    childrenKey := fmt.Sprintf("agent:%s:children", agentID)
    as.redis.Del(childrenKey)

    return nil
}

func (as *AgentStore) GetChildren(parentID string) ([]string, error) {
    childrenKey := fmt.Sprintf("agent:%s:children", parentID)
    return as.redis.LRange(childrenKey, 0, -1)
}

func (as *AgentStore) GetActiveAgents() (map[string]string, error) {
    return as.redis.HGetAll("agents:active")
}

func (as *AgentStore) Close() error {
    return as.redis.Close()
}