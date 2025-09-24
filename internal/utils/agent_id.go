package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func CreatePrimaryAgentID() string {
	return fmt.Sprintf("agent:%s", uuid.New().String())
}

func CreateSubAgentID(primaryAgentID string) string {
	primaryUUID := strings.TrimPrefix(primaryAgentID, "agent:")
	return fmt.Sprintf("agent:%s:%s", primaryUUID, uuid.New().String())
}

func IsPrimaryAgent(agentID string) bool {
	parts := strings.Split(agentID, ":")
	return len(parts) == 2 && parts[0] == "agent"
}

func IsSubAgent(agentID string) bool {
	parts := strings.Split(agentID, ":")
	return len(parts) == 3 && parts[0] == "agent"
}

func GetPrimaryAgentID(subAgentID string) (string, error) {
	if !IsSubAgent(subAgentID) {
		return "", fmt.Errorf("not a sub-agent: %s", subAgentID)
	}

	parts := strings.Split(subAgentID, ":")
	return fmt.Sprintf("agent:%s", parts[1]), nil
}
