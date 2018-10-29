package message

import (
	"sync"

	"github.com/gladiusio/legion/utils"
	flatbuffers "github.com/google/flatbuffers/go"
)

// builderPool keeps a dynamically sized pool of builder objects
var builderPool = sync.Pool{
	New: func() interface{} {
		return flatbuffers.NewBuilder(0)
	},
}

// New returns a new Message with the specified fields
func New(sender *utils.LegionAddress, messageType string, body []byte) *Message {
	return &Message{}
}

// Message is the message type that the network expects.
// It is simple by design to allow for signifigant customization
// of the network and it's message processing.
type Message struct {
	sender      *utils.LegionAddress
	messageType string
	body        []byte
}

// Sender returns the sender address
func (m *Message) Sender() *utils.LegionAddress {
	return m.sender
}

// Type gets the message type
func (m *Message) Type() string {
	return m.messageType
}

// Body returns arbitrary body bytes, could be another flatbuffer
func (m *Message) Body() []byte {
	return m.body
}

// Checksum returns the message checksum
func (m *Message) Checksum() []byte {
	return []byte{}
}

// Encode encodes the data as a byte array
func (m *Message) Encode() []byte {
	// Get a cached or new builder, then reset it. This is to limit allocations
	b := builderPool.Get().(*flatbuffers.Builder)
	b.Reset()

	return []byte{}
}

// UnMarshal populates the message wiht fields by unpacking the buffer
func (m *Message) Decode(buf []byte) error {
	return nil
}
