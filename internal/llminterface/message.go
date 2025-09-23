package llminterface

const (
	MessageTypeUser       = "user"
	MessageTypeSystem     = "system"
	MessageTypeAssistant  = "assistant"
	MessageTypeToolCall   = "tool_call"
	MessageTypeToolResult = "tool_result"
)

type Message struct {
	Type       string         `json:"type"`
	Content    string         `json:"content,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolName   string         `json:"tool_name,omitempty"`
	Arguments  map[string]any `json:"arguments,omitempty"`
	Result     string         `json:"result,omitempty"`
}

func (m Message) GetType() string { return m.Type }

type RequestMessageList []Message
type ResponseMessageList []Message
type MessageList []Message

func NewUserMessage(content string) Message {
	return Message{Type: MessageTypeUser, Content: content}
}

func NewSystemMessage(content string) Message {
	return Message{Type: MessageTypeSystem, Content: content}
}

func NewAssistantMessage(content string) Message {
	return Message{Type: MessageTypeAssistant, Content: content}
}

func NewToolCallMessage(toolCallID, toolName string, arguments map[string]any) Message {
	return Message{Type: MessageTypeToolCall, ToolCallID: toolCallID, ToolName: toolName, Arguments: arguments}
}

func NewToolResultMessage(toolCallID, toolName, result string) Message {
	return Message{Type: MessageTypeToolResult, ToolCallID: toolCallID, ToolName: toolName, Result: result}
}
