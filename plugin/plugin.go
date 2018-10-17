package plugin

// Interface is an inteface that allows an implementer to interact with
// the network in various ways described in the methods below.
type Interface interface {
	NewMessage(ctx *MessageContext)
	PeerAdded(ctx *PeerContext)
	PeerPromotion(ctx *PeerContext)
	PeerDeleted(ctx *PeerContext)
	Startup(ctx *NetworkContext)
	Close(ctx *NetworkContext)
}

// Generic is a type used to expose methods so a plugin doesn't need
// to have all of the required methods
type Generic struct{}

// NewMessage is called when a message is recieve by the network
func (*Generic) NewMessage(ctx *MessageContext) {}

// PeerAdded is called when a peer is added to the network
func (*Generic) PeerAdded(ctx *PeerContext) {}

// PeerPromotion is called when a peer is made writeable
func (*Generic) PeerPromotion(ctx *PeerContext) {}

// PeerDeleted is called when a peer is deleted
func (*Generic) PeerDeleted(ctx *PeerContext) {}

// Startup is called when the local peer starts listening
func (*Generic) Startup(ctx *NetworkContext) {}

// Close is called when the network is shutdown
func (*Generic) Close(ctx *NetworkContext) {}
