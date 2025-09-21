package llminterface

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "time"
)

type OpenAIClient struct {
    apiKey     string
    apiBase    string
    httpClient *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
    return &OpenAIClient{
        apiKey:  apiKey,
        apiBase: "https://api.openai.com/v1",
        httpClient: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

func (c *OpenAIClient) CreateChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
    // Default model if not specified
    if request.Model == "" {
        request.Model = "gpt-4"
    }

    // Prepare request body
    requestBody, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // Create HTTP request
    url := fmt.Sprintf("%s/chat/completions", c.apiBase)
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

    // Log request details
    log.Printf("OpenAI API request - Model: %s, Messages: %d, Tools: %d",
        request.Model, len(request.Messages), len(request.Tools))

    // Make request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Check for errors
    if resp.StatusCode != http.StatusOK {
        var errorResp struct {
            Error struct {
                Message string `json:"message"`
                Type    string `json:"type"`
                Code    string `json:"code"`
            } `json:"error"`
        }
        
        if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
            return nil, fmt.Errorf("OpenAI API error (%s): %s", errorResp.Error.Type, errorResp.Error.Message)
        }
        
        return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
    }

    // Parse response
    var response ChatResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    // Log response details
    log.Printf("OpenAI API response - ID: %s, Model: %s, Tokens: %d",
        response.ID, response.Model, response.Usage.TotalTokens)

    if len(response.Choices) > 0 && len(response.Choices[0].Message.ToolCalls) > 0 {
        log.Printf("OpenAI API response includes %d tool calls", 
            len(response.Choices[0].Message.ToolCalls))
    }

    return &response, nil
}

// SetAPIBase allows overriding the API base URL (for testing or using proxies)
func (c *OpenAIClient) SetAPIBase(apiBase string) {
    c.apiBase = apiBase
}

// SetHTTPClient allows setting a custom HTTP client
func (c *OpenAIClient) SetHTTPClient(client *http.Client) {
    c.httpClient = client
}