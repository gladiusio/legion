package message

import "github.com/gladiusio/legion/utils"

// Interface is the message interface that the network expects.
// It is simple by design to allow for signifigant customization
// of the network and it's message processing.
type Interface interface {
	Sender() utils.LegionAddress
	Type() string
	Body() []byte // Arbitrary body bytes, could be another flatbuffer
	Checksum() []byte
	Data() []byte // Data that is needed outside the body, for example details about compression of the body
	Marshal() []byte
	UnMarshal() error
}
