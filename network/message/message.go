package message

import (
	"errors"
	"sync"

	"github.com/gladiusio/legion/network/message/flatb"
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
func New(sender utils.LegionAddress, messageType string, body []byte) *Message {
	return &Message{
		sender:      sender,
		messageType: messageType,
		body:        body,
	}
}

// Message is the message type that the network expects.
// It is simple by design to allow for significant customization
// of the network and it's message processing.
type Message struct {
	sender      utils.LegionAddress
	messageType string
	body        []byte
}

// Sender returns the sender address
func (m *Message) Sender() utils.LegionAddress {
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

// Encode encodes the data as a byte array
func (m *Message) Encode() []byte {
	b := builderPool.Get().(*flatbuffers.Builder) // Get a cached or new builder
	defer builderPool.Put(b)                      // Return it to the pool when we're done

	// Reset the builder
	b.Reset()

	// Build our fields
	body := b.CreateByteString(m.body)
	sender := b.CreateString(m.sender.String())
	messageType := b.CreateString(m.messageType)

	// Set the fields
	flatb.MessageStart(b)
	flatb.MessageAddBody(b, body)
	flatb.MessageAddSender(b, sender)
	flatb.MessageAddType(b, messageType)

	// End the message
	end := flatb.MessageEnd(b)
	b.Finish(end)

	return b.FinishedBytes()
}

// Decode populates the message with fields by unpacking the buffer
func (m *Message) Decode(buf []byte) error {
	decoded := flatb.GetRootAsMessage(buf, 0)

	if decoded.BodyBytes() == nil || decoded.Sender() == nil || decoded.Type() == nil {
		return errors.New("message: error unpacking buffer")
	}

	m.body = decoded.BodyBytes()
	m.sender = utils.LegionAddressFromString(string(decoded.Sender()))
	m.messageType = string(decoded.Type())

	return nil
}
