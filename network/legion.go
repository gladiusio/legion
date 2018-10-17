package network

import (
	"sync"

	"github.com/gladiusio/legion/message"
	"github.com/gladiusio/legion/plugin"
	"github.com/gladiusio/legion/utils"
	multierror "github.com/hashicorp/go-multierror"
)

// Legion is a type with methods to interface with the network
type Legion struct {
	// These peers are written to by the broadcast and broadcast random (when a peer list isn't provided).
	// They are generally considered safe to write and read to, and have been explicitly
	// protmoted by calling PromotePeer(address)
	promotedPeers *sync.Map

	// This is the initial state of a peer when it is added, it allows communication
	// with plugins (useful for a peer authorization step), but will not be written to unless
	// specifically called
	allPeers *sync.Map

	// Registered plugins, these are called in order when plugin events happen
	plugins []plugin.Interface
}

// Broadcast sends the message to all writeable peers, unless a
// specified list of peers is provided
func (l *Legion) Broadcast(message *message.Message, addresses ...utils.KCPAddress) {
	// Send to all promoted peers
	if len(addresses) == 0 {
		l.promotedPeers.Range(func(k, v interface{}) bool { v.(*Peer).QueueMessage(message); return true })
		return
	}

	// If they provided a set of addresses we can check from all connected peers (not just promoted)
	for _, address := range addresses {
		if p, ok := l.allPeers.Load(address); ok {
			p.(*Peer).QueueMessage(message)
		}
	}
}

// BroadcastRandom broadcasts a message to N random writeable peers
func (l *Legion) BroadcastRandom(message *message.Message, n int) {

}

// AddPeer adds the specified peer(s) to the network by dialing it and
// opening a stream, as well as adding it to the list of all peers.
// Note: this does not add it to the promoted peers, so a broadcast
// to all peers will not send to the added peers unless they are
// promoted
func (l *Legion) AddPeer(addresses ...utils.KCPAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		p := NewPeer(address)
		err := p.Dial()
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		l.allPeers.Store(address, p)
	}
	return result.ErrorOrNil()
}

// PromotePeer makes the given peer(s) writeable, if the peer doesn't exist
// it is created first.
func (l *Legion) PromotePeer(addresses ...utils.KCPAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		// If the peer exists, we add it to the promoted peers
		if p, ok := l.allPeers.Load(address); ok {
			l.promotedPeers.Store(address, p)
		} else {
			p := NewPeer(address)
			err := p.Dial()
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			l.allPeers.Store(address, p)
			l.promotedPeers.Store(address, p)
		}
	}
	return result.ErrorOrNil()
}

// DeletePeer closes all connections to a peer(s) and removes it from all peer lists
func (l *Legion) DeletePeer(address ...utils.KCPAddress) error {
	return nil
}

// PeerExists returns whether or not a peer has been connected to previously
func (l *Legion) PeerExists(address utils.KCPAddress) bool {
	return false
}

// PeerPromoted returns whether or not a peer is promoted.
func (l *Legion) PeerPromoted(address utils.KCPAddress) bool {
	return false
}

// RegisterPlugin registers a plugin(s) with the network
func (l *Legion) RegisterPlugin(p ...*plugin.Interface) {

}

// Listen will listen on the configured address for incomming connections, it will
// also wait for all plugin's Startup() methods to return before binding.
func (l *Legion) Listen() {
	// Setup start and stop calls on our plugin list
	l.FireNetworkEvent(StartupEvent)
	defer l.FireNetworkEvent(CloseEvent)

	// TODO: Listen loop goes here
}

// FireMessageEvent fires a new message event and sends context to the correct plugin
// methods based on the event type
func (l *Legion) FireMessageEvent(eventType MessageEvent, message *message.Message) {
	go func() {
		messageContext := &plugin.MessageContext{} // Create some context for our plugin
		for _, p := range l.plugins {
			if eventType == NewMessageEvent {
				go p.NewMessage(messageContext)
			}
		}
	}()
}

// FirePeerEvent fires a peer event and sends context to the correct plugin methods
// based on the event type
func (l *Legion) FirePeerEvent(eventType PeerEvent, peer ...*Peer) {
	go func() {
		peerContext := &plugin.PeerContext{} // Create some context for our plugin
		for _, p := range l.plugins {
			if eventType == PeerAddEvent {
				go p.PeerAdded(peerContext)
			} else if eventType == PeerDeleteEvent {
				go p.PeerDeleted(peerContext)
			} else if eventType == PeerPromotionEvent {
				go p.PeerPromotion(peerContext)
			}
		}
	}()
}

// FireNetworkEvent fires a network event and sends network context to the correct
// plugin method based on the event type. NOTE: This method blocks until all are
// completed
func (l *Legion) FireNetworkEvent(eventType NetEvent) {
	netContext := &plugin.NetworkContext{}
	for _, p := range l.plugins {
		if eventType == StartupEvent {
			p.Startup(netContext)
		} else if eventType == CloseEvent {
			p.Close(netContext)
		}
	}
}
