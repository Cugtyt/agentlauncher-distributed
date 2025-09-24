package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TaskRequest struct {
	Task         string `json:"task"`
	SystemPrompt string `json:"system_prompt"`
}

type TaskResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

type TaskResult struct {
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
}

func main() {
	baseURL := "http://localhost:8080"

	fmt.Println("🚀 Testing Agent Launcher Distributed System")
	fmt.Println("==================================================")

	// Test 1: Health check
	fmt.Println("\n1. Testing health endpoint...")
	if err := testHealth(baseURL); err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	fmt.Println("✅ Health check passed")

	// Test 2: Calculator tool
	fmt.Println("\n2. Testing calculator tool...")
	calcTaskID, err := testCalculator(baseURL)
	if err != nil {
		fmt.Printf("❌ Calculator test failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Calculator task created: %s\n", calcTaskID)

	// Test 3: Current time tool
	fmt.Println("\n3. Testing current time tool...")
	timeTaskID, err := testCurrentTime(baseURL)
	if err != nil {
		fmt.Printf("❌ Time test failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Time task created: %s\n", timeTaskID)

	// Test 4: Random number tool
	fmt.Println("\n4. Testing random number tool...")
	randomTaskID, err := testRandomNumber(baseURL)
	if err != nil {
		fmt.Printf("❌ Random number test failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Random number task created: %s\n", randomTaskID)

	// Wait for processing
	fmt.Println("\n5. Waiting for tasks to complete...")
	time.Sleep(10 * time.Second)

	// Check results
	fmt.Println("\n6. Checking task results...")

	tasks := []struct {
		name string
		id   string
	}{
		{"Calculator", calcTaskID},
		{"Time", timeTaskID},
		{"Random", randomTaskID},
	}

	for _, task := range tasks {
		fmt.Printf("\n📋 Checking %s task (%s):\n", task.name, task.id)
		result, err := getTaskResult(baseURL, task.id)
		if err != nil {
			fmt.Printf("❌ Failed to get result: %v\n", err)
			continue
		}

		fmt.Printf("  Status: %s\n", result.Status)
		if result.Result != "" {
			fmt.Printf("  Result: %s\n", result.Result)
		}

		if result.Status == "success" {
			fmt.Printf("✅ %s task completed successfully\n", task.name)
		} else if result.Status == "failed" {
			fmt.Printf("❌ %s task failed\n", task.name)
		} else {
			fmt.Printf("⏳ %s task still pending\n", task.name)
		}
	}

	fmt.Println("\n🎉 Test suite completed!")
}

func testHealth(baseURL string) error {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(body) != "OK" {
		return fmt.Errorf("unexpected health response: %s", string(body))
	}

	return nil
}

func testCalculator(baseURL string) (string, error) {
	task := TaskRequest{
		Task:         "Calculate 15 + 25 and tell me the result",
		SystemPrompt: "You are a helpful assistant that can use tools to help users with calculations.",
	}

	return createTask(baseURL, task)
}

func testCurrentTime(baseURL string) (string, error) {
	task := TaskRequest{
		Task:         "What is the current time?",
		SystemPrompt: "You are a helpful assistant that can tell users the current time.",
	}

	return createTask(baseURL, task)
}

func testRandomNumber(baseURL string) (string, error) {
	task := TaskRequest{
		Task:         "Generate a random number between 1 and 100",
		SystemPrompt: "You are a helpful assistant that can generate random numbers for users.",
	}

	return createTask(baseURL, task)
}

func createTask(baseURL string, task TaskRequest) (string, error) {
	jsonData, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task: %v", err)
	}

	resp, err := http.Post(baseURL+"/tasks", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create task: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("task creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		return "", fmt.Errorf("failed to decode task response: %v", err)
	}

	return taskResp.AgentID, nil
}

func getTaskResult(baseURL, agentID string) (*TaskResult, error) {
	resp, err := http.Get(baseURL + "/results?agent_id=" + agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &TaskResult{
			AgentID: agentID,
			Status:  "not_found",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get result with status %d: %s", resp.StatusCode, string(body))
	}

	var result TaskResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode result: %v", err)
	}

	return &result, nil
}
