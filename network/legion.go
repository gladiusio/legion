package network

import (
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/events"
	"github.com/gladiusio/legion/network/transport"
	"github.com/gladiusio/legion/utils"

	log "github.com/gladiusio/legion/logger"

	multierror "github.com/hashicorp/go-multierror"
)

// NewLegion creates a legion object from a config
func NewLegion(conf *config.LegionConfig, f Framework) *Legion {
	if f == nil {
		log.Warn().Log("legion: using generic framework for validation and introductions")
		f = &GenericFramework{}
	}
	return &Legion{
		peers:     &sync.Map{},
		plugins:   make([]PluginInterface, 0),
		config:    conf,
		started:   make(chan struct{}),
		framework: f,
	}
}

// Legion is a type with methods to interface with the network
type Legion struct {
	// All connected peers stored as: [LegionAddress -> Peer]
	peers *sync.Map

	// Which framework legion is using
	framework Framework

	// Registered plugins, these are called in order when plugin events happen
	plugins []PluginInterface

	// Our config type
	config *config.LegionConfig

	// started is a channel that blocks until Listen() completes
	started chan struct{}

	// listener is the network listener that legion listens for new connections on
	listener net.Listener
}

// Me returns the local bindaddress
func (l *Legion) Me() utils.LegionAddress {
	return l.config.AdvertiseAddress
}

// Broadcast sends the message to all writeable peers, unless a
// specified list of peers is provided
func (l *Legion) Broadcast(message *transport.Message, addresses ...utils.LegionAddress) {
	// Wait until we're listening
	l.Started()

	// Send to all peers
	if len(addresses) == 0 {
		l.peers.Range(func(k, v interface{}) bool { v.(*Peer).QueueMessage(message); return true })
		return
	}

	// If they provided addresses, we can send to those
	for _, address := range addresses {
		if p, ok := l.peers.Load(address); ok {
			p.(*Peer).QueueMessage(message)
		} else {
			l.AddPeer(address)
		}
	}
}

// BroadcastRandom broadcasts a message to N random promoted peers
func (l *Legion) BroadcastRandom(message *transport.Message, n int) {
	// Wait until we're listening
	l.Started()

	// sync.Map doesn't store length, so we get n random like this
	addrs := make([]utils.LegionAddress, 0, 100)
	l.peers.Range(func(key, value interface{}) bool { addrs = append(addrs, key.(utils.LegionAddress)); return true })

	if n > len(addrs) || n <= 1 {
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
func (l *Legion) AddPeer(addresses ...utils.LegionAddress) error {
	var result *multierror.Error
	for _, address := range addresses {
		// Make sure the peer isn't already added or ourselves
		if _, ok := l.peers.Load(address); !ok && address != l.Me() {
			p, err := l.createAndDialPeer(address)
			if err != nil {
				log.Warn().Field("err", err).Log("Error adding peer")
				result = multierror.Append(result, err)
				continue
			}
			l.storePeer(p)
			l.addMessageListener(p, false)
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
		if p, ok := l.peers.Load(address); ok {
			err := p.(*Peer).Close()
			if err != nil {
				result = multierror.Append(result, err)
			}
			l.peers.Delete(address)
		}
	}

	return result.ErrorOrNil()
}

// PeerExists returns whether or not a peer has been connected to previously
func (l *Legion) PeerExists(address utils.LegionAddress) bool {
	_, ok := l.peers.Load(address)
	return ok
}

// DoAllPeers runs the function f on all peers
func (l *Legion) DoAllPeers(f func(p *Peer)) {
	l.peers.Range(func(key, value interface{}) bool {
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
	// Configure our framework
	l.framework.Configure(l)

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
func (l *Legion) FireMessageEvent(eventType events.MessageEvent, message *transport.Message) {
	go func() {
		messageContext := &MessageContext{Legion: l, Message: message, Sender: utils.LegionAddressFromString(message.GetSender())} // Create some context for our plugin
		for _, p := range l.plugins {
			if eventType == events.NewMessageEvent {
				go p.NewMessage(messageContext)
			}
		}
	}()
}

// FirePeerEvent fires a peer event and sends context to the correct plugin methods
// based on the event type
func (l *Legion) FirePeerEvent(eventType events.PeerEvent, peer *Peer, isIncoming bool) {
	go func() {
		// Create some context for our plugin
		peerContext := &PeerContext{
			Legion:     l,
			Peer:       peer,
			IsIncoming: isIncoming,
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
				l.storePeer(p)
				l.FirePeerEvent(events.PeerAddEvent, p, true)
			}
		}
		// Get our reveive channel
		receiveChan := p.IncomingMessages()
		for {
			select {
			case m, open := <-receiveChan:
				if !open {
					return
				}

				// Call the framework validator to see if the message should be sent to plugins
				ctx := &MessageContext{Legion: l, Message: m, Sender: utils.LegionAddressFromString(m.GetSender())}
				if l.framework.ValidateMessage(ctx) {
					l.FireMessageEvent(events.NewMessageEvent, m)
					// Only store the peer on the first message if it is an incoming connection
					// this is so we can get a the actual sender and store it
					if incoming {
						once.Do(storePeer(utils.LegionAddressFromString(m.GetSender())))
					} else {
						once.Do(func() { l.FirePeerEvent(events.PeerAddEvent, p, false) })
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

	// Introduce ourselves
	l.framework.Introduce(l, p)

	return p, nil
}

// NewMessage returns a message with the sender field set to the bind address of the network and no extra data
func (l *Legion) NewMessage(messageType string, body []byte) *transport.Message {
	return &transport.Message{Type: messageType, Body: body, Sender: l.Me().String()}
}

func (l *Legion) handleNewConnection(conn net.Conn) {
	// Create a new peer that's not yet stored, it will be registered with the network
	// later in the message listener once we recieve a message from it.
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

func (l *Legion) storePeer(p *Peer) {
	// If it is already stored, don't add a cleanup handler
	if _, stored := l.peers.LoadOrStore(p.remote, p); stored {
		return
	}

	// Handle cleanup - wait until that peer is disconnected to remove it
	go func() {
		p.BlockUntilDisconnected()

		p.Close()

		// Cleanup the peer map
		l.peers.Delete(p.remote)

		l.FirePeerEvent(events.PeerDisconnectEvent, p, true)
		log.Debug().Field("remote_addr", p.Remote().String()).Log("Peer disconnected")
	}()
}
