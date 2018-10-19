package network

import (
	"bufio"
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

	rw *bufio.ReadWriter
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

// OpenStream dials the remote and opens a stream to the peer
func (p *Peer) OpenStream() error {
	// Get a TCP connection
	conn, err := net.Dial("tcp", p.remote.String())
	if err != nil {
		return err
	}

	// Setup client side of yamux
	session, err := yamux.Client(conn, nil)
	if err != nil {
		return err
	}

	// Open a new stream
	stream, err := session.Open()
	if err != nil {
		return err
	}

	p.rw = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	return nil
}

// RecieveStream takes an incoming connection and creates a stream from it
func (p *Peer) RecieveStream(conn net.Conn) error {
	// Setup server side of yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		return err
	}

	// Accept a stream
	stream, err := session.Accept()
	if err != nil {
		return err
	}

	p.rw = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	return nil
}

// Close closes the stream if it exists
func (p *Peer) Close() error {
	return nil
}
