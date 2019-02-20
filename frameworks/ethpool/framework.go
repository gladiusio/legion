package ethpool

import (
	"github.com/gladiusio/legion/network"
)

// IncomingMessage represents an incoming message after parsing
type IncomingMessage struct {
	Sender string
	Body   []byte
}

// New returns a Framework that uses the specified function to check if an address is valid, if
// nil all addresses will be considered valid
func New(addressValidator func(string) bool) *Framework {
	return &Framework{
		addressValidator,
	}
}

// Framework is a framework for interacting with other peers using ethereum signatures and a kademlia style DHT,
// and only accepting messages from peers that are specified as valid.
type Framework struct {
	// Used to check if an address is acceptable
	addressValidator func(string) bool

	// TODO: Ethereum key for signing and address

	// TODO: Add system for getting state messages
}

// Configure is used to set anything up you want with Legion (loading plugins etc),
// it is called at startup
func (f *Framework) Configure(l *network.Legion) {
	// Register our DHT plugin
	dht := new(DHT)
	l.RegisterPlugin(dht)

	// Potentially register plugins todo with state
}

// ValidateMessage is called before any message is passed to plugins
func (f *Framework) ValidateMessage(ctx *network.MessageContext) bool { return true }

// Introduce is called a peer is connected to (add or recieve)
func (f *Framework) Introduce(l *network.Legion, p *network.Peer) {
	p.QueueMessage(l.NewMessage("legion_introduction", []byte{}))
}

// SendMessage will send a signed version of the message to specified recipient
// it will error if the recipient can't be connected to or found
func (f *Framework) SendMessage(recipient, messageType string, body []byte) error {
	return nil
}

// RecieveMessageChan returns a channel that recieves messages of the specified type
func (f *Framework) RecieveMessageChan(messageType string) chan *IncomingMessage {
	return nil
}
