package message

import (
	"encoding/json"
	"time"
)

// Message is a type that stores timestamp and content and can return the json
// serialized version
type Message struct {
	Content   *json.RawMessage `json:"content"`
	Timestamp int64            `json:"timestamp"`
}

// New creates a new Message type with fields for timestamp and a json message
func New(jsonMessage []byte) *Message {
	h := json.RawMessage(jsonMessage)
	return &Message{Content: &h, Timestamp: time.Now().Unix()}
}

func NewBlankMessage() *Message {
	return &Message{Content: nil, Timestamp: time.Now().Unix()}
}

// Serialize returns a serialized JSON string that includes the current timestamp
func (m Message) Serialize() []byte {
	bytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return bytes
}
