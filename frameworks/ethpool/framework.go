package ethpool

import (
	"github.com/gladiusio/legion/network"
)

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
func (*Framework) Configure(l *network.Legion) {
	// Build DHT plugin

	// Register our DHT
	dht := new(DHT)
	l.RegisterPlugin(dht)

	// Potentially register plugins todo with state
}

// ValidateMessage is called before any message is passed to plugins
func (*Framework) ValidateMessage(ctx *network.MessageContext) bool { return true }

// Introduce is called a peer is connected to (add or recieve)
func (*Framework) Introduce(l *network.Legion, p *network.Peer) {
	p.QueueMessage(l.NewMessage("legion_introduction", []byte{}))
}
