package network

import (
	"github.com/gladiusio/legion/network/transport"
	"github.com/gladiusio/legion/utils"

	"errors"
)

// MessageContext has context for a given message such as the legion object
// and methods to interact with the remote peer that sent the message
type MessageContext struct {
	Sender  utils.LegionAddress
	Message *transport.Message
	Legion  *Legion
}

// Reply is a helper method to reply to an incoming message
func (mc *MessageContext) Reply(msg *transport.Message) error {
	// If this is an RPC message we should send a reply, if not just send a regular message
	if mc.Message.IsRequest {
		p, exists := mc.Legion.peers.Load(mc.Sender)
		if exists {
			p.(*Peer).QueueReply(mc.Message.RpcId, msg)
		} else {
			return errors.New("legion: error sending reply to peer")
		}
	} else {
		mc.Legion.Broadcast(msg, mc.Sender)
	}

	return nil
}

// PeerContext has context for a peer event such as the legion object and
// the peer change that fired the event
type PeerContext struct {
	Legion     *Legion
	Peer       *Peer
	IsIncoming bool
}

// NetworkContext is general context of the network, gives access to just
// the legion object and a few other helpers
type NetworkContext struct {
	Legion *Legion
}
