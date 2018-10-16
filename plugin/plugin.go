package plugin

// Plugin is an inteface that allows an implementer to interact with the network
// in various ways described in the methods below.
type Plugin interface {
	// When a new message is recieved by the network, NewMessage() is
	// called on all registered plugins
	NewMessage()

	// When a new peer is added to the network via the AddPeer() method,
	// PeerAdded() is called on all registered plugins
	PeerAdded()

	// When a peer is promoted to writeable by the PromotePeer() method,
	// PeerPromotion is called on all registered plugins
	PeerPromotion()

	// When a peer is deleted from the network by the DeletePeer() method,
	// PeerDeleted is called on all registered plugins
	PeerDeleted()

	// Startup() and Close() are called at network startup and close
	// repsectively
	Startup()
	Close()
}
