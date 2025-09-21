package llminterface

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
)

type LLMHandler func(messages RequestMessageList, tools RequestToolList, agentid string, eventbus eventbus.EventBus) ResponseMessageList