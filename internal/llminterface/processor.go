package llminterface

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
)

type LLMProcessor func(messages RequestMessageList, tools RequestToolList, agentid string, eventbus eventbus.EventBus) (ResponseMessageList, error)
