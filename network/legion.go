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
		} else {
			l.AddPeer(address)
		}
	}
}

// BroadcastRandom broadcasts a message to N random promoted peers
func (l *Legion) BroadcastRandom(message *message.Message, n int) {
	// sync.Map doesn't store length, so we get n random like this
	addrs := make([]utils.KCPAddress, 0, 100)
	l.promotedPeers.Range(func(key, value interface{}) bool { addrs = append(addrs, key.(utils.KCPAddress)); return true })

	// addrs contains all peers addresses now, lets select N random
	for i := 0; i < n; i++ {
		l.Broadcast(message, addrs[i])
		addrs = append(addrs[:i], addrs[i+1:]...) // Cut out i to stop repeat broadcasts
	}
}

// AddPeer adds the specified peer(s) to the network by dialing it and
// opening a stream, as well as adding it to the list of all peers.
// Note: this does not add it to the promoted peers, so a broadcast
// to all peers will not send to the added peers unless they are
// promoted. Returns an error if one or more peers can't be dialed,
// however all peers will have a dial attempt.
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
		l.addMessageListener(p)
	}
	return result.ErrorOrNil()
}

// PromotePeer makes the given peer(s) writeable, if the peer doesn't exist
// it is created first. Returns an error if one or more peers can't be dialed,
// however all peers will have a dial attempt.
func (l *Legion) PromotePeer(addresses ...utils.KCPAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		if p, ok := l.allPeers.Load(address); ok { // If the peer exists, we add it to the promoted peers
			l.promotedPeers.Store(address, p.(*Peer))
			l.FirePeerEvent(PeerPromotionEvent, p.(*Peer))
		} else { // If not we create a new peer and dial it
			p := NewPeer(address)
			err := p.Dial()
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			l.allPeers.Store(address, p)
			l.promotedPeers.Store(address, p)
			l.FirePeerEvent(PeerPromotionEvent, p)
		}
	}
	return result.ErrorOrNil()
}

// DeletePeer closes all connections to a peer(s) and removes it from all peer lists.
// Returns an error if there is an error closing one or more peers. No matter the
// error, there will be an attempt to close all peers.
func (l *Legion) DeletePeer(addresses ...utils.KCPAddress) {
	for _, address := range addresses {
		if p, ok := l.allPeers.Load(address); ok {
			p.(*Peer).Close()
			l.allPeers.Delete(address)
			l.promotedPeers.Delete(address)
		}
	}
}

// PeerExists returns whether or not a peer has been connected to previously
func (l *Legion) PeerExists(address utils.KCPAddress) bool {
	_, ok := l.allPeers.Load(address)
	return ok
}

// PeerPromoted returns whether or not a peer is promoted.
func (l *Legion) PeerPromoted(address utils.KCPAddress) bool {
	_, ok := l.promotedPeers.Load(address)
	return ok
}

// RegisterPlugin registers a plugin(s) with the network
func (l *Legion) RegisterPlugin(plugins ...plugin.Interface) {
	for _, p := range plugins {
		l.plugins = append(l.plugins, p)
	}
}

// Listen will listen on the configured address for incomming connections, it will
// also wait for all plugin's Startup() methods to return before binding.
func (l *Legion) Listen() {
	// Setup start and stop calls on our plugin list
	l.FireNetworkEvent(StartupEvent)
	defer l.FireNetworkEvent(CloseEvent)

	// TODO: Listen loop goes here, this would see an incoming steam and
	// create a peer in l.allPeers

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

func (l *Legion) addMessageListener(p *Peer) {
	// Listen to messages from the peer forever
	go func() {
		for {
			select {
			case m := <-p.IncomingMessages():
				l.FireMessageEvent(NewMessageEvent, m)
			}
		}
	}()
}
