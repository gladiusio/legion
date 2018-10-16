package plugin

// Plugin is an inteface that allows an implementer to interact with the network
// in various ways described in the methods below.
type Plugin interface {
	// When a new message is recieved by the network, NewMessage() is
	// called on all registered plugins
	NewMessage(ctx *PluginContext)

	// When a new peer is added to the network via the AddPeer() method,
	// PeerAdded() is called on all registered plugins
	PeerAdded(ctx *PeerContext)

	// When a peer is promoted to writeable by the PromotePeer() method,
	// PeerPromotion is called on all registered plugins
	PeerPromotion(ctx *PeerContext)

	// When a peer is deleted from the network by the DeletePeer() method,
	// PeerDeleted is called on all registered plugins
	PeerDeleted(ctx *PeerContext)

	// Startup() and Close() are called at network startup and close
	// repsectively
	Startup(ctx *NetworkContext)
	Close(ctx *NetworkContext)
}

// MessageContext has context for a given message such as the legion object
// and methods to interact with the remote peer that sent the message
type MessageContext struct {
}

// PeerContext has context for a peer event such as the legion object and
// the peer change that fired the event
type PeerContext struct {
}

// NetworkContext is general context of the network, gives access to just
// the legion object and a few other helpers
type NetworkContext struct {
}
