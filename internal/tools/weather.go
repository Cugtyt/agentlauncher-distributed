package tools

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"
)

type WeatherTool struct{}

func NewWeatherTool() *WeatherTool {
    return &WeatherTool{}
}

func (t *WeatherTool) Name() string {
    return "weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather information for a location"
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    location, ok := args["location"].(string)
    if !ok {
        return nil, fmt.Errorf("location argument is required and must be a string")
    }

    log.Printf("WeatherTool: Getting weather for '%s'", location)

    // Simulate weather API call
    // In production, this would call a real weather API (OpenWeatherMap, etc.)
    weather := t.mockWeather(location)

    return weather, nil
}

func (t *WeatherTool) mockWeather(location string) map[string]interface{} {
    // Generate mock weather data
    // In production, replace with actual weather API call
    
    rand.Seed(time.Now().UnixNano())
    
    // Base temperature varies by location name (simple mock logic)
    baseTemp := 20.0
    if strings.Contains(strings.ToLower(location), "north") {
        baseTemp = 5.0
    } else if strings.Contains(strings.ToLower(location), "south") {
        baseTemp = 25.0
    }
    
    temp := baseTemp + rand.Float64()*10 - 5
    humidity := 40 + rand.Intn(40)
    windSpeed := 5 + rand.Float64()*15
    
    conditions := []string{"Clear", "Partly Cloudy", "Cloudy", "Light Rain", "Overcast"}
    condition := conditions[rand.Intn(len(conditions))]
    
    return map[string]interface{}{
        "location": location,
        "current": map[string]interface{}{
            "temperature_c": fmt.Sprintf("%.1f", temp),
            "temperature_f": fmt.Sprintf("%.1f", temp*9/5+32),
            "condition":     condition,
            "humidity":      humidity,
            "wind_kph":      fmt.Sprintf("%.1f", windSpeed),
            "wind_mph":      fmt.Sprintf("%.1f", windSpeed*0.621371),
            "feels_like_c":  fmt.Sprintf("%.1f", temp-2),
            "feels_like_f":  fmt.Sprintf("%.1f", (temp-2)*9/5+32),
            "uv_index":      rand.Intn(11),
            "visibility_km": 10 + rand.Intn(20),
        },
        "forecast": map[string]interface{}{
            "tomorrow": map[string]interface{}{
                "high_c":    fmt.Sprintf("%.1f", temp+3),
                "low_c":     fmt.Sprintf("%.1f", temp-2),
                "condition": conditions[rand.Intn(len(conditions))],
            },
        },
        "timestamp": time.Now().Format(time.RFC3339),
    }
}