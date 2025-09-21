package events

const (
	TaskCreateEventName   = "task.create"
	TaskFinishEventName   = "task.finish"
	AgentCreateEventName  = "agent.create"
	AgentStartEventName   = "agent.start"
	AgentFinishEventName  = "agent.finish"
	AgentDeletedEventName = "agent.deleted"
	LLMRequestEventName   = "llm.request"
	LLMResponseEventName  = "llm.response"
	ToolExecuteEventName  = "tool.execute"
	ToolStartEventName    = "tool.start"
	ToolFinishEventName   = "tool.finish"
	ToolResultEventName   = "tool.result"
	MessageAddEventName   = "message.add"
	MessageGetEventName   = "message.get"
)
