package websocket

import "encoding/json"

const (
	MessageTypeNotification = "notification"
	MessageTypeChatMessage  = "chat.message"
	MessageTypeTyping       = "chat.typing"
)

// Message is sent to connected clients over WebSocket.
type Message struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// Encode serializes a WebSocket message.
func Encode(msg Message) ([]byte, error) {
	return json.Marshal(msg)
}
