package events

const (
	AgentCreateEventName       = "agent-create"
	AgentStartEventName        = "agent-start"
	AgentFinishEventName       = "agent-finish"
	AgentErrorEventName        = "agent-error"
	AgentRuntimeErrorEventName = "agent-runtime-error"
	AgentDeletedEventName      = "agent-deleted"

	TaskCreateEventName = "task-create"
	TaskFinishEventName = "task-finish"
	TaskErrorEventName  = "task-error"

	LLMRequestEventName  = "llm-request"
	LLMResponseEventName = "llm-response"
	LLMErrorEventName    = "llm-error"

	MessageAddEventName         = "message-add"
	MessageStreamStartEventName = "message-stream-start"
	MessageStreamDeltaEventName = "message-stream-delta"
	MessageStreamDoneEventName  = "message-stream-done"
	MessageStreamErrorEventName = "message-stream-error"

	ToolCallStreamNameEventName      = "toolcall-stream-name"
	ToolCallStreamArgsStartEventName = "toolcall-stream-args-start"
	ToolCallStreamArgsDeltaEventName = "toolcall-stream-args-delta"
	ToolCallStreamArgsDoneEventName  = "toolcall-stream-args-done"
	ToolCallStreamArgsErrorEventName = "toolcall-stream-args-error"

	ToolExecRequestEventName  = "tool-exec-request"
	ToolExecResultsEventName  = "tool-exec-results"
	ToolExecStartEventName    = "tool-exec-start"
	ToolExecFinishEventName   = "tool-exec-finish"
	ToolExecErrorEventName    = "tool-exec-error"
	ToolRuntimeErrorEventName = "tool-runtime-error"
)
