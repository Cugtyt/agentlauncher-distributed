package tools

import (
    "context"
    "fmt"
    "log"
    "time"
)

type SearchTool struct{}

func NewSearchTool() *SearchTool {
    return &SearchTool{}
}

func (t *SearchTool) Name() string {
    return "search"
}

func (t *SearchTool) Description() string {
    return "Search for information on the web"
}

func (t *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    query, ok := args["query"].(string)
    if !ok {
        return nil, fmt.Errorf("query argument is required and must be a string")
    }

    log.Printf("SearchTool: Searching for '%s'", query)

    // Simulate search with mock results
    // In production, this would call a real search API (Google, Bing, etc.)
    results := t.mockSearch(query)

    return map[string]interface{}{
        "query":      query,
        "results":    results,
        "total":      len(results),
        "timestamp":  time.Now().Format(time.RFC3339),
    }, nil
}

func (t *SearchTool) mockSearch(query string) []map[string]string {
    // Mock search results based on query
    // In production, replace with actual search API call
    
    mockResults := []map[string]string{
        {
            "title":   fmt.Sprintf("Search result 1 for: %s", query),
            "snippet": fmt.Sprintf("This is a relevant result about %s with detailed information.", query),
            "url":     fmt.Sprintf("https://example.com/search?q=%s&result=1", query),
        },
        {
            "title":   fmt.Sprintf("Search result 2 for: %s", query),
            "snippet": fmt.Sprintf("Another important finding related to %s from a trusted source.", query),
            "url":     fmt.Sprintf("https://example.com/search?q=%s&result=2", query),
        },
        {
            "title":   fmt.Sprintf("Search result 3 for: %s", query),
            "snippet": fmt.Sprintf("Additional context and information about %s for comprehensive understanding.", query),
            "url":     fmt.Sprintf("https://example.com/search?q=%s&result=3", query),
        },
    }

    // Add some variety based on query content
    if len(query) > 20 {
        mockResults = append(mockResults, map[string]string{
            "title":   "Extended search result",
            "snippet": "For longer queries, we provide more comprehensive results with additional context.",
            "url":     "https://example.com/extended",
        })
    }

    return mockResults
}