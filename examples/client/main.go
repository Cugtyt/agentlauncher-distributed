package main

import (
    "fmt"
    "log"
    "time"

    "github.com/yourusername/agentlauncher-distributed/pkg/api"
)

func main() {
    // Create API client
    client := api.NewClient("http://localhost:8080")

    // Check health
    health, err := client.GetHealth()
    if err != nil {
        log.Printf("Warning: Health check failed: %v", err)
    } else {
        log.Printf("Service health: %+v", health)
    }

    // Example 1: Simple task
    simpleTask := api.TaskRequest{
        Task: "What is the weather in New York?",
        Tools: []string{"weather"},
    }

    fmt.Println("\n=== Creating simple task ===")
    response, err := client.CreateTask(simpleTask)
    if err != nil {
        log.Fatalf("Failed to create task: %v", err)
    }

    fmt.Printf("Task created successfully!\n")
    fmt.Printf("Agent ID: %s\n", response.AgentID)
    fmt.Printf("Status: %s\n", response.Status)
    fmt.Printf("Message: %s\n", response.Message)

    // Example 2: Complex task with sub-agents
    complexTask := api.TaskRequest{
        Task: "Research and compare weather in three cities: New York, London, and Tokyo. Then search for travel tips for the city with the best weather.",
        Tools: []string{"weather", "search", "create_agent"},
        Metadata: map[string]string{
            "request_id": "example-001",
            "user_id":    "demo-user",
        },
    }

    fmt.Println("\n=== Creating complex task with sub-agents ===")
    response2, err := client.CreateTask(complexTask)
    if err != nil {
        log.Fatalf("Failed to create complex task: %v", err)
    }

    fmt.Printf("Complex task created successfully!\n")
    fmt.Printf("Agent ID: %s\n", response2.AgentID)
    fmt.Printf("Status: %s\n", response2.Status)

    // Example 3: Task with context
    contextTask := api.TaskRequest{
        Task:    "Find flights from San Francisco to Tokyo next month",
        Context: "User prefers business class, direct flights, and morning departures. Budget is flexible.",
        Tools:   []string{"search"},
    }

    fmt.Println("\n=== Creating task with context ===")
    response3, err := client.CreateTask(contextTask)
    if err != nil {
        log.Fatalf("Failed to create task with context: %v", err)
    }

    fmt.Printf("Context task created successfully!\n")
    fmt.Printf("Agent ID: %s\n", response3.AgentID)

    fmt.Println("\n=== All tasks created ===")
    fmt.Println("Check the logs to see the agents processing:")
    fmt.Println("kubectl logs -n agentlauncher -l app=agent-runtime -f")
}