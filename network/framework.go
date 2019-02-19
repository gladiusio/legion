package network

// Framework is an interface that allows you to modify the underlying communication
// of legion
type Framework interface {
	// Set anything up you want with Legion when the Listen method is called (loading plugins etc).
	// Should block until the framework is ready to accept messages.
	Configure(*Legion)

	// Called before any message is passed to plugins
	ValidateMessage(*MessageContext) bool

	// When a peer is dialed Introduce is called
	Introduce(*Legion, *Peer)
}

// GenericFramework is a type used to expose methods so a framework doesn't need
// to have all of the required methods (it is also used as the default framework)
type GenericFramework struct{}

// Configure is used to set anything up you want with Legion (loading plugins etc),
// it is called at startup
func (*GenericFramework) Configure(*Legion) {}

// ValidateMessage is called before any message is passed to plugins
func (*GenericFramework) ValidateMessage(ctx *MessageContext) bool { return true }

// Introduce is called a peer is connected to (add or recieve)
func (*GenericFramework) Introduce(l *Legion, p *Peer) {
	p.QueueMessage(l.NewMessage("legion_introduction", []byte{}))
}
