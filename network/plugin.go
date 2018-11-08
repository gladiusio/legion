package network

// PluginInterface is an inteface that allows an implementer to interact with
// the network in various ways described in the methods below.
type PluginInterface interface {
	NewMessage(ctx *MessageContext)
	PeerAdded(ctx *PeerContext)
	PeerPromotion(ctx *PeerContext)
	PeerDisconnect(ctx *PeerContext)
	Startup(ctx *NetworkContext)
	Close(ctx *NetworkContext)
}

// GenericPlugin is a type used to expose methods so a plugin doesn't need
// to have all of the required methods
type GenericPlugin struct{}

// NewMessage is called when a message is received by the network
func (*GenericPlugin) NewMessage(ctx *MessageContext) {}

// PeerAdded is called when a peer is added to the network
func (*GenericPlugin) PeerAdded(ctx *PeerContext) {}

// PeerPromotion is called when a peer is made writeable
func (*GenericPlugin) PeerPromotion(ctx *PeerContext) {}

// PeerDisconnect is called when a peer is deleted
func (*GenericPlugin) PeerDisconnect(ctx *PeerContext) {}

// Startup is called when the local peer starts listening
func (*GenericPlugin) Startup(ctx *NetworkContext) {}

// Close is called when the network is shutdown
func (*GenericPlugin) Close(ctx *NetworkContext) {}
