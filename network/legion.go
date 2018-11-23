package network

import (
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/events"
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"

	log "github.com/gladiusio/legion/logger"

	multierror "github.com/hashicorp/go-multierror"
)

// NewLegion creates a legion object from a config
func NewLegion(conf *config.LegionConfig) *Legion {
	if conf.MessageValidator == nil {
		log.Warn().Log("legion: message validator function is nil, all messages will be considered valid")
		conf.MessageValidator = func(m *message.Message) bool { return true }
	}
	return &Legion{
		promotedPeers: &sync.Map{},
		allPeers:      &sync.Map{},
		plugins:       make([]PluginInterface, 0),
		config:        conf,
		started:       make(chan struct{}),
	}
}

// Legion is a type with methods to interface with the network
type Legion struct {
	// These peers are written to by the broadcast and broadcast random (when a peer list isn't provided).
	// They are generally considered safe to write and read to, and have been explicitly
	// promoted by calling PromotePeer(address)
	promotedPeers *sync.Map

	// This is the initial state of a peer when it is added, it allows communication
	// with plugins (useful for a peer authorization step), but will not be written to unless
	// specifically called
	allPeers *sync.Map

	// Registered plugins, these are called in order when plugin events happen
	plugins []PluginInterface

	// Our config type
	config *config.LegionConfig

	// started is a channel that blocks until Listen() completes
	started chan struct{}

	// listener is the network listener that legion listens for new connections on
	listener net.Listener
}

// Broadcast sends the message to all writeable peers, unless a
// specified list of peers is provided
func (l *Legion) Broadcast(message *message.Message, addresses ...utils.LegionAddress) {
	// Wait until we're listening
	l.Started()

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
	// Wait until we're listening
	l.Started()

	// sync.Map doesn't store length, so we get n random like this
	addrs := make([]utils.LegionAddress, 0, 100)
	l.promotedPeers.Range(func(key, value interface{}) bool { addrs = append(addrs, key.(utils.LegionAddress)); return true })

	if n > len(addrs) {
		l.Broadcast(message)
		return
	}

	// addrs contains all peers addresses now, lets select N random
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		rn := r.Intn(len(addrs) - 1)
		l.Broadcast(message, addrs[rn])
		addrs = append(addrs[:rn], addrs[rn+1:]...) // Cut out i to stop repeat broadcasts
	}
}

// AddPeer adds the specified peer(s) to the network by dialing it and
// opening a stream, as well as adding it to the list of all peers.
// Note: this does not add it to the promoted peers, so a broadcast
// to all peers will not send to the added peers unless they are
// promoted. Returns an error if one or more peers can't be dialed,
// however all peers will have a dial attempt.
func (l *Legion) AddPeer(addresses ...utils.LegionAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		if _, ok := l.allPeers.Load(address); !ok { // Make sure the peer isn't already added
			p, err := l.createAndDialPeer(address)
			if err != nil {
				log.Warn().Field("err", err).Log("Error adding peer")
				result = multierror.Append(result, err)
				continue
			}
			l.storePeer(p, false)
			l.addMessageListener(p, false)
		}
	}
	return result.ErrorOrNil()
}

// PromotePeer makes the given peer(s) writeable, if the peer doesn't exist
// it is created first. Returns an error if one or more peers can't be dialed,
// however all peers will have a dial attempt.
func (l *Legion) PromotePeer(addresses ...utils.LegionAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		if p, ok := l.allPeers.Load(address); ok { // If the peer exists, we add it to the promoted peers
			l.storePeer(p.(*Peer), true)
			l.FirePeerEvent(events.PeerPromotionEvent, p.(*Peer))
		} else { // If not we create a new peer and dial it
			p, err := l.createAndDialPeer(address)
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			l.storePeer(p, true)
			l.addMessageListener(p, false)
			l.FirePeerEvent(events.PeerAddEvent, p)
			l.FirePeerEvent(events.PeerPromotionEvent, p)
		}
	}
	return result.ErrorOrNil()
}

// DeletePeer closes all connections to a peer(s) and removes it from all peer lists.
// Returns an error if there is an error closing one or more peers. No matter the
// error, there will be an attempt to close all peers.
func (l *Legion) DeletePeer(addresses ...utils.LegionAddress) error {
	var result *multierror.Error

	for _, address := range addresses {
		if p, ok := l.allPeers.Load(address); ok {
			err := p.(*Peer).Close()
			if err != nil {
				result = multierror.Append(result, err)
			}
			l.allPeers.Delete(address)
			l.promotedPeers.Delete(address)
		}
	}

	return result.ErrorOrNil()
}

// PeerExists returns whether or not a peer has been connected to previously
func (l *Legion) PeerExists(address utils.LegionAddress) bool {
	_, ok := l.allPeers.Load(address)
	return ok
}

// PeerPromoted returns whether or not a peer is promoted.
func (l *Legion) PeerPromoted(address utils.LegionAddress) bool {
	_, ok := l.promotedPeers.Load(address)
	return ok
}

// DoAllPeers runs the function f on all peers
func (l *Legion) DoAllPeers(f func(p *Peer)) {
	l.allPeers.Range(func(key, value interface{}) bool {
		p, ok := value.(*Peer)
		if ok {
			f(p)
		}
		return true
	})
}

// DoPromotedPeers runs the function f on all promoted peers
func (l *Legion) DoPromotedPeers(f func(p *Peer)) {
	l.promotedPeers.Range(func(key, value interface{}) bool {
		p, ok := value.(*Peer)
		if ok {
			f(p)
		}
		return true
	})
}

// RegisterPlugin registers a plugin(s) with the network
func (l *Legion) RegisterPlugin(plugins ...PluginInterface) {
	for _, p := range plugins {
		l.plugins = append(l.plugins, p)
	}
}

// Listen will listen on the configured address for incoming connections, it will
// also wait for all plugin's Startup() methods to return before binding.
func (l *Legion) Listen() error {
	var err error

	l.listener, err = net.Listen("tcp", l.config.BindAddress.String())
	if err != nil {
		return err
	}

	// Signal after a delay we're listening
	go func() {
		time.Sleep(1 * time.Second)
		close(l.started)
		log.Info().Field("addr", l.config.BindAddress.String()).Log("Listening on: " + l.config.BindAddress.String())
		l.FireNetworkEvent(events.StartupEvent)
	}()

	// Accept incoming TCP connections
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			continue
		}

		// Handle the incoming connection and create a peer
		go l.handleNewConnection(conn)
	}
}

// Stop closes the listener and fires the plugin stop event
func (l *Legion) Stop() error {
	defer l.FireNetworkEvent(events.CloseEvent)
	return l.listener.Close()
}

// Started blocks until the network is running
func (l *Legion) Started() {
	<-l.started
}

// FireMessageEvent fires a new message event and sends context to the correct plugin
// methods based on the event type
func (l *Legion) FireMessageEvent(eventType events.MessageEvent, message *message.Message) {
	go func() {
		messageContext := &MessageContext{Legion: l, Message: message} // Create some context for our plugin
		for _, p := range l.plugins {
			if eventType == events.NewMessageEvent {
				go p.NewMessage(messageContext)
			}
		}
	}()
}

// FirePeerEvent fires a peer event and sends context to the correct plugin methods
// based on the event type
func (l *Legion) FirePeerEvent(eventType events.PeerEvent, peer *Peer) {
	go func() {
		// Create some context for our plugin
		peerContext := &PeerContext{
			Legion: l,
			Peer:   peer,
		}
		// Tell all of the plugins about the event
		for _, p := range l.plugins {
			if eventType == events.PeerAddEvent {
				go p.PeerAdded(peerContext)
			} else if eventType == events.PeerDisconnectEvent {
				go p.PeerDisconnect(peerContext)
			} else if eventType == events.PeerPromotionEvent {
				go p.PeerPromotion(peerContext)
			}
		}
	}()
}

// FireNetworkEvent fires a network event and sends network context to the correct
// plugin method based on the event type. NOTE: This method blocks until all are
// completed
func (l *Legion) FireNetworkEvent(eventType events.NetworkEvent) {
	netContext := &NetworkContext{Legion: l}
	for _, p := range l.plugins {
		if eventType == events.StartupEvent {
			p.Startup(netContext)
		} else if eventType == events.CloseEvent {
			p.Close(netContext)
		}
	}
}

func (l *Legion) addMessageListener(p *Peer, incoming bool) {
	// Listen to messages from the peer forever
	go func() {
		var once sync.Once
		storePeer := func(sender utils.LegionAddress) func() {
			return func() {
				p.remote = sender
				l.storePeer(p, false)
				l.FirePeerEvent(events.PeerAddEvent, p)
			}
		}
		for {
			select {
			case m := <-p.IncomingMessages():
				// Call whatever validator is registered to see if the message is valid
				if l.config.MessageValidator(m) {
					l.FireMessageEvent(events.NewMessageEvent, m)
					// Only store the peer on the first message if it is an incoming connection
					// this is so we can get a the actual sender and store it
					if incoming {
						once.Do(storePeer(m.Sender()))
					}
				}
			}
		}
	}()
}

func (l *Legion) createAndDialPeer(address utils.LegionAddress) (*Peer, error) {
	p := NewPeer(address)

	conn, err := net.Dial("tcp", address.String())
	if err != nil {
		return nil, err
	}

	err = p.CreateClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Send an introduction message so it knows who we are
	p.QueueMessage(l.NewMessage("legion_introduction", []byte{}))

	return p, nil
}

// NewMessage returns a message with the sender field set to the bind address of the network
func (l *Legion) NewMessage(messageType string, body []byte) *message.Message {
	return message.New(l.config.BindAddress, messageType, body)
}

func (l *Legion) handleNewConnection(conn net.Conn) {
	// Create a new (not stored or dialable) peer. This will be registered with the network
	// later in the message listener.
	p := NewPeer(utils.LegionAddress{})
	err := p.CreateServer(conn)
	if err != nil {
		conn.Close()
		return
	}

	// Listen to new messages from that peer
	l.addMessageListener(p, true)

	log.Debug().Field("addr", conn.RemoteAddr().String()).Log("Received new peer connection")
}

func (l *Legion) storePeer(p *Peer, promoted bool) {
	if promoted {
		l.promotedPeers.Store(p.remote, p)
	}

	// If it is already stored, don't add a cleanup handler
	if _, stored := l.allPeers.LoadOrStore(p.remote, p); stored {
		return
	}

	// Wait until that peer is disconnected to remove it
	go func() {
		p.BlockUntilDisconnected()

		// Cleanup both maps
		l.allPeers.Delete(p.remote)
		l.promotedPeers.Delete(p.remote)

		// Only fire this once (not in storePromotedPeer)
		l.FirePeerEvent(events.PeerDisconnectEvent, p)
		log.Debug().Field("remote_addr", p.Remote().String()).Log("Peer disconnected")
	}()
}
