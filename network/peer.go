package network

import (
	"net"

	"github.com/gladiusio/legion/utils"
	"github.com/hashicorp/yamux"
)

// NewPeer returns a new peer from the given remote. It also
// sets up the reading and writing channels
func NewPeer(remote utils.LegionAddress) *Peer {
	return &Peer{remote: remote}
}

// Peer is an type that allows easy communication with
// a remote peer
type Peer struct {
	// The remote Address to dial
	remote utils.LegionAddress

	// The internal channel we write to to send a new message
	// to the remote
	sendQueue chan *Message

	// The channel of incoming messages
	recieveChan chan *Message

	session *yamux.Session
}

// QueueMessage queues the specified message to be sent to the remote
func (p *Peer) QueueMessage(m *Message) {
	go func() { p.sendQueue <- m }()
}

// IncomingMessages returns a channel of every message recieved from
// the remote peer
func (p *Peer) IncomingMessages() chan *Message {
	return p.recieveChan
}

// CreateSession takes an incoming connection and creates a session from
func (p *Peer) CreateSession(conn net.Conn) error {
	// Setup server side of yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		return err
	}

	// Store this session so we can open streams and write messages to it
	p.session = session
	return nil
}

// Close closes the stream if it exists
func (p *Peer) Close() error {
	return p.session.Close()
}

func (p *Peer) startSendLoop() {
	go func() {
		for {
			select {
			case m := <-p.sendQueue:
				go p.sendMessage(m)
			}
		}
	}()
}

func (p *Peer) sendMessage(m *Message) {
	stream, err := p.session.Open()
	if err != nil {
		// TODO: Log error
	}

	messageBytes := []byte{}

	stream.Write(messageBytes)
}

func (p *Peer) startRecieveLoop() {
	go func() {
		for {
			incomingStream, err := p.session.Accept()
			if err != nil {
				// TODO: Log error
			}

			go p.readMessage(incomingStream)
		}
	}()
}

func (p *Peer) readMessage(conn net.Conn) {
	// TODO: Parse the message from the conn then close the stream

	m := &Message{}

	go func() { p.recieveChan <- m }()
}
