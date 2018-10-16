package network

import (
	"github.com/gladiusio/legion/message"
	"github.com/gladiusio/legion/plugin"
	"github.com/gladiusio/legion/utils"
)

// Legion is a type with methods to interface with the network
type Legion struct {
	plugins []plugin.Interface // Registered plugins, these are called when plugin events happen

}

// Broadcast sends the message to all writeable peers, unless a
// specified list of peers is provided
func (l *Legion) Broadcast(message *message.Message, peer ...*Peer) {

}

// BroadcastRandom broadcasts a message to N random writeable peers
func (l *Legion) BroadcastRandom(message *message.Message, n int) {

}

// AddPeer adds the specified peer(s) to the network by opening a stream
// and storing it in the open streams list. This however does not authorize
// a stream to be writeable by the Broadcast methods, that must be done by
// marking it as writable by calling PromotePeer()
func (l *Legion) AddPeer(address ...utils.KCPAddress) {

}

// PromotePeer makes the given peer(s) writeable
func (l *Legion) PromotePeer(address ...utils.KCPAddress) {

}

// DeletePeer closes all connections to a peer(s) and removes it from all peer lists
func (l *Legion) DeletePeer(address ...utils.KCPAddress) {

}

// RegisterPlugin registers a plugin(s) with the network
func (l *Legion) RegisterPlugin(p ...*plugin.Interface) {

}

// Listen will listen on the configured address for incomming connections, it will
// also wait for all plugin's Startup() methods to return before binding.
func (l *Legion) Listen() {

}

// FireMessageEvent fires a new message event and sends context to the correct plugin
// methods based on the event type
func (l *Legion) FireMessageEvent(eventType MessageEvent, message *message.Message) error {
	return nil
}

// FirePeerEvent fires a peer event and sends context to the correct plugin methods
// based on the event type
func (l *Legion) FirePeerEvent(eventType PeerEvent, peer ...*Peer) error {
	return nil
}

// FireNetworkEvent fires a network event and sends network context to the correct
// plugin method based on the event type
func (l *Legion) FireNetworkEvent(eventType NetEvent) error {
	return nil
}
