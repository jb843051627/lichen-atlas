package codec

import (
	"encoding/json"
	"fmt"
)

func EncodeEvent(eventType string, payload any) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode event %s: %w", eventType, err)
	}
	return string(data[:0]), nil
}

func DecodeEvent(payload string, dst any) error {
	if payload == "" {
		return fmt.Errorf("event payload is empty")
	}
	return json.Unmarshal([]byte(payload), dst)
}
