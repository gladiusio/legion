package network

// Framework is an interface that allows you to modify the underlying communication
// of legion
type Framework interface {
	// Set anything up you want with Legion when the Listen method is called.
	// Should block until the framework is ready to accept messages.
	Configure(*Legion) error

	// Called before any message is passed to plugins
	ValidateMessage(*MessageContext) bool

	// Methods to interact with legion
	NewMessage(*MessageContext)
	PeerAdded(*PeerContext)
	PeerDisconnect(*PeerContext)
	Startup(*NetworkContext)
	Close(*NetworkContext)
}

// GenericFramework is a type used to expose methods so a framework doesn't need
// to have all of the required methods (it is also used as the default framework)
type GenericFramework struct{}

// Configure is used to set anything up you want with Legion (loading plugins etc),
// it is called at startup
func (*GenericFramework) Configure(*Legion) error { return nil }

// ValidateMessage is called before any message is passed to plugins
func (*GenericFramework) ValidateMessage(ctx *MessageContext) bool { return true }

// NewMessage is called when a message is received by the network
func (*GenericFramework) NewMessage(ctx *MessageContext) {}

// PeerAdded is called when a peer is added to the network
func (*GenericFramework) PeerAdded(ctx *PeerContext) {}

// PeerDisconnect is called when a peer is deleted
func (*GenericFramework) PeerDisconnect(ctx *PeerContext) {}

// Startup is called when the local peer starts listening
func (*GenericFramework) Startup(ctx *NetworkContext) {}

// Close is called when the network is shutdown
func (*GenericFramework) Close(ctx *NetworkContext) {}
