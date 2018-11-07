package network

import "github.com/gladiusio/legion/network/message"

// MessageContext has context for a given message such as the legion object
// and methods to interact with the remote peer that sent the message
type MessageContext struct {
	Message *message.Message
	Legion  *Legion
}

// Reply is a helper method to reply to an incoming message
func (mc *MessageContext) Reply(msg *message.Message) {
	mc.Legion.Broadcast(msg, msg.Sender())
}

// PeerContext has context for a peer event such as the legion object and
// the peer change that fired the event
type PeerContext struct {
	Legion *Legion
	Peer   *Peer
}

// NetworkContext is general context of the network, gives access to just
// the legion object and a few other helpers
type NetworkContext struct {
	Legion *Legion
}
