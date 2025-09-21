package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Client is the API client for the agent launcher
type Client struct {
    baseURL    string
    httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// CreateTask creates a new task and returns the agent ID
func (c *Client) CreateTask(request TaskRequest) (*TaskResponse, error) {
    // Marshal request
    body, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // Create HTTP request
    url := fmt.Sprintf("%s/tasks", c.baseURL)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    // Make request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    // Read response
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Check status
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
    }

    // Parse response
    var taskResp TaskResponse
    if err := json.Unmarshal(respBody, &taskResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return &taskResp, nil
}

// GetHealth checks the health of the service
func (c *Client) GetHealth() (*HealthStatus, error) {
    url := fmt.Sprintf("%s/health", c.baseURL)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("service unhealthy: status %d", resp.StatusCode)
    }

    var health HealthStatus
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        // If it's just a simple OK response
        return &HealthStatus{
            Service:   "agent-launcher",
            Status:    "healthy",
            Timestamp: time.Now(),
        }, nil
    }

    return &health, nil
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
    c.httpClient.Timeout = timeout
}

// SetHTTPClient sets a custom HTTP client
func (c *Client) SetHTTPClient(client *http.Client) {
    c.httpClient = client
}