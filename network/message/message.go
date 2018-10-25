package message

import "github.com/gladiusio/legion/utils"

// New returns a new Message with the specified data
func New(sender utils.LegionAddress, messageType string, body, checksum []byte, compression bool) *Message {
	m := &Message{
		Sender:      sender,
		Type:        messageType,
		Body:        body,
		Compression: compression,
	}

	m.calculateChecksum()

	return m
}

// Message is the message type that the network expects.
// It is simple by design to allow for significant customization
// of the network and it's message processing.
type Message struct {
	Sender      utils.LegionAddress `json:"sender"`
	Type        string              `json:"type"`
	Body        []byte              `json:"body"`
	Checksum    []byte              `json:"checksum"`
	Compression bool                `json:"compression"`
}

func (m *Message) calculateChecksum() {
	// TODO: make this actually calculate checksum
	m.Checksum = []byte{}
}
