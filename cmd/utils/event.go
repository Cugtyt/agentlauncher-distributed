package utils

import (
	"encoding/json"
	"log"
)

func UnmarshalEvent[T any](data []byte, eventName string) (T, bool) {
	var event T
	if err := json.Unmarshal(data, &event); err != nil {
		log.Printf("Failed to unmarshal %s event: %v", eventName, err)
		return event, false
	}
	return event, true
}
