package network

import (
	"bufio"
	"encoding/binary"
	"net"

	"github.com/gladiusio/legion/logger"
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
	"github.com/hashicorp/yamux"
)

// NewPeer returns a new peer from the given remote. It also
// sets up the reading and writing channels
func NewPeer(remote utils.LegionAddress) *Peer {
	p := &Peer{
		remote:       remote,
		sendQueue:    make(chan *message.Message),
		receiveChans: make([](chan (*message.Message)), 0, 1),
	}

	return p
}

// Peer is an type that allows easy communication with
// a remote peer
type Peer struct {
	// The remote Address to dial
	remote utils.LegionAddress

	// The internal channel we write to to send a new message
	// to the remote
	sendQueue chan *message.Message

	// The channel of incoming messages
	receiveChans [](chan *message.Message)

	// The session with the remote (either incoming or outgoing)
	session *yamux.Session

	// The disconnected channel, closed when the peer disconnects
	disconnected chan struct{}
}

// QueueMessage queues the specified message to be sent to the remote
func (p *Peer) QueueMessage(m *message.Message) {
	go func() { p.sendQueue <- m }()
}

// IncomingMessages registers a new listen channel and returns it
func (p *Peer) IncomingMessages() chan *message.Message {
	r := make(chan *message.Message)
	p.receiveChans = append(p.receiveChans, r)
	return r
}

// BlockUntilDisconnected blocks until the remote is disconnected
func (p *Peer) BlockUntilDisconnected() {
	<-p.session.CloseChan()
}

// CreateClient takes an incoming connection and creates a client session from it
func (p *Peer) CreateClient(conn net.Conn) error {
	// Setup client side of yamux
	session, err := yamux.Client(conn, nil)
	if err != nil {
		return err
	}

	// Store this session so we can open streams and write messages to it
	p.session = session

	p.startSendLoop()
	p.startRecieveLoop()

	return nil
}

// CreateServer takes an incoming connection and creates a server session from it
func (p *Peer) CreateServer(conn net.Conn) error {
	// Setup server side of yamux
	session, err := yamux.Server(conn, nil)
	if err != nil {
		return err
	}

	// Store this session so we can open streams and write messages to it
	p.session = session

	p.startSendLoop()
	p.startRecieveLoop()

	return nil
}

// Close closes the stream if it exists
func (p *Peer) Close() error {
	for _, c := range p.receiveChans {
		close(c)
	}
	close(p.sendQueue)

	return p.session.Close()
}

// Remote returns the address of the remote peer
func (p *Peer) Remote() utils.LegionAddress {
	return p.remote
}

func (p *Peer) startSendLoop() {
	go func() {
		for {
			select {
			case m, open := <-p.sendQueue:
				if p.session.IsClosed() || !open {
					return
				}
				go p.sendMessage(m)
			}
		}
	}()
}

func (p *Peer) sendMessage(m *message.Message) {
	stream, err := p.session.Open()
	if err != nil {
		logger.Warn().Field("err", err.Error()).Log("peer: error opening connection")
		return
	}

	messageBytes := m.Encode()

	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(len(messageBytes)))

	buffer = append(buffer, messageBytes...)

	bw := bufio.NewWriter(stream)
	_, err = bw.Write(buffer)
	if err != nil {
		logger.Warn().Field("err", err.Error()).Log("peer: error writing to stream")
		return
	}
	bw.Flush()
}

func (p *Peer) startRecieveLoop() {
	go func() {
		for {
			incomingStream, err := p.session.AcceptStream()
			if err != nil {
				if p.session.IsClosed() {
					return
				}
				logger.Warn().Field("err", err.Error()).Log("peer: error establishing incoming stream from peer.")
				p.session.Close()
				return
			}

			go p.readMessage(incomingStream)
		}
	}()
}

func (p *Peer) readMessage(stream *yamux.Stream) {
	var err error
	buffer := make([]byte, 4)

	// Close this message stream when we're done
	defer stream.Close()

	// Read the message size header
	numBytesRead := 0
	for numBytesRead < 4 {
		n, err := stream.Read(buffer)
		if err != nil {
			logger.Debug().Field("err", err).Log("Error reading message header")
			return
		}

		numBytesRead += n
	}
	// Convert it into an int
	size := binary.BigEndian.Uint32(buffer)
	if size == 0 || size > 1e+8 {
		return
	}

	// Allocate the message size
	buffer = make([]byte, size)

	// Read into the buffer
	numBytesRead = 0
	for numBytesRead < int(size) {
		n, err := stream.Read(buffer[numBytesRead:]) // Make sure we don't write over anthing
		if err != nil {
			logger.Debug().Field("err", err).Log("Error reading message")
			return
		}
		numBytesRead += n
	}
	// Unmarshal the message
	m := &message.Message{}
	err = m.Decode(buffer)
	if err != nil {
		logger.Debug().Field("remote_peer", stream.RemoteAddr().String()).Log("peer: could not decode incoming message")
		return
	}

	// Send off our message into the receive chans
	for _, rchan := range p.receiveChans {
		go func(c chan *message.Message) { c <- m }(rchan)
	}

}
