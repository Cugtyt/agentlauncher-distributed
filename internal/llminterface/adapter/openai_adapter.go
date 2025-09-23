package adapter

import (
	"encoding/json"

	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

func ConvertMessagesToOpenAI(messages llminterface.RequestMessageList) []map[string]interface{} {
	openaiMessages := make([]map[string]interface{}, 0)
	var currentToolCalls []map[string]interface{}

	for i, msg := range messages {
		switch msg.Type {
		case llminterface.MessageTypeUser:
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, map[string]interface{}{
					"role":       "assistant",
					"tool_calls": currentToolCalls,
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, map[string]interface{}{
				"role":    "user",
				"content": msg.Content,
			})

		case llminterface.MessageTypeAssistant:
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, map[string]interface{}{
					"role":       "assistant",
					"tool_calls": currentToolCalls,
				})
				currentToolCalls = nil
			}

			hasToolCalls := false
			if i+1 < len(messages) {
				hasToolCalls = messages[i+1].Type == llminterface.MessageTypeToolCall
			}

			if !hasToolCalls {
				openaiMessages = append(openaiMessages, map[string]interface{}{
					"role":    "assistant",
					"content": msg.Content,
				})
			}

		case llminterface.MessageTypeSystem:
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, map[string]interface{}{
					"role":       "assistant",
					"tool_calls": currentToolCalls,
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, map[string]interface{}{
				"role":    "system",
				"content": msg.Content,
			})

		case llminterface.MessageTypeToolCall:
			argsBytes, _ := json.Marshal(msg.Arguments)
			currentToolCalls = append(currentToolCalls, map[string]interface{}{
				"id":   msg.ToolCallID,
				"type": "function",
				"function": map[string]interface{}{
					"name":      msg.ToolName,
					"arguments": string(argsBytes),
				},
			})

		case llminterface.MessageTypeToolResult:
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, map[string]interface{}{
					"role":       "assistant",
					"tool_calls": currentToolCalls,
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, map[string]interface{}{
				"role":         "tool",
				"content":      msg.Result,
				"tool_call_id": msg.ToolCallID,
			})
		}
	}

	if len(currentToolCalls) > 0 {
		openaiMessages = append(openaiMessages, map[string]interface{}{
			"role":       "assistant",
			"tool_calls": currentToolCalls,
		})
	}

	return openaiMessages
}

func ConvertToolsToOpenAI(tools llminterface.RequestToolList) []map[string]interface{} {
	openaiTools := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		parameters := make(map[string]interface{})
		required := []string{}

		for _, param := range tool.Parameters {
			parameters[param.Name] = map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}
			if param.Type == "array" && param.Items != nil {
				parameters[param.Name].(map[string]interface{})["items"] = param.Items
			}
			if param.Required {
				required = append(required, param.Name)
			}
		}

		openaiTools[i] = map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": parameters,
					"required":   required,
				},
			},
		}
	}

	return openaiTools
}

func ConvertOpenAIResponseToMessages(content string, toolCalls []map[string]interface{}) llminterface.ResponseMessageList {
	response := llminterface.ResponseMessageList{}

	if content != "" {
		response = append(response, llminterface.NewAssistantMessage(content))
	}

	for _, toolCall := range toolCalls {
		if function, ok := toolCall["function"].(map[string]interface{}); ok {
			var args map[string]any
			if argsStr, ok := function["arguments"].(string); ok {
				json.Unmarshal([]byte(argsStr), &args)
			}

			response = append(response, llminterface.NewToolCallMessage(
				toolCall["id"].(string),
				function["name"].(string),
				args,
			))
		}
	}

	return response
}
