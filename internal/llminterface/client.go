package llminterface

import "context"

// LLMClient interface for different LLM providers
type LLMClient interface {
    CreateChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error)
}