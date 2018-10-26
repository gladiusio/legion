package message

import (
	"github.com/gladiusio/legion/utils"
	flatbuffers "github.com/google/flatbuffers/go"
)

// New returns a new Message with the specified fields
func New() *Message {
	return nil
}

// Message is the message type that the network expects.
// It is simple by design to allow for signifigant customization
// of the network and it's message processing.
type Message struct {
	b *flatbuffers.Builder
}

// Sender returns the sender address
func (m *Message) Sender() utils.LegionAddress {

}

// Type gets the message type
func (m *Message) Type() string {

}

// Body returns arbitrary body bytes, could be another flatbuffer
func (m *Message) Body() []byte {

}

// Checksum returns the message checksum
func (m *Message) Checksum() []byte {

}

// Data returns data that is needed outside the body,
// for example details about compression of the body
func (m *Message) Data() []byte {

}

// Marshal marshals the data as a byte array
func (m *Message) Marshal() []byte {
	return []byte{}
}

// UnMarshal popultes the type by unpacking the buffer
func (m *Message) UnMarshal() error {
	return nil
}
