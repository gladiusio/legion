package network

import "github.com/gladiusio/legion/utils"

// Message is the message interface that the network expects.
// It is simple by design to allow for signifigant customization
// of the network and it's message processing.
type Message struct {
	Sender      utils.KCPAddress
	Type        string
	Body        []byte
	Checksum    []byte
	Compression bool
}
