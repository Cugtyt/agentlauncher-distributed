package llminterface

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
)

type LLMProcessor func(messages []Message, tools RequestToolList, agentid string, eventbus eventbus.EventBus) ([]Message, error)
