package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gladiusio/legion/logger"
	"github.com/gladiusio/legion/network/transport"
	"github.com/gladiusio/legion/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/yamux"
	"go.uber.org/atomic"
)

// NewPeer returns a new peer from the given remote. It also
// sets up the reading and writing channels
func NewPeer(remote utils.LegionAddress) *Peer {
	p := &Peer{
		remote:      remote,
		sendQueue:   make(chan *transport.Message),
		receiveChan: make(chan (*transport.Message)),
		requests:    make(map[uint64]chan *transport.Message),
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
	sendQueue chan *transport.Message

	// The channel of incoming messages
	receiveChan chan *transport.Message

	// The session with the remote (either incoming or outgoing)
	session *yamux.Session

	// The current RPC id we're using
	rcpID atomic.Uint64

	// Stores the RPC requests by ID
	requests    map[uint64]chan *transport.Message // Uint64 -> chan *transport.Message
	requestsMux sync.Mutex

	// The disconnected channel, closed when the peer disconnects
	disconnected chan struct{}
}

type logWriter struct{}

func (*logWriter) Write(p []byte) (int, error) {
	logger.Debug().Field("yamux_message", string(p))
	return len(p), nil
}

// QueueMessage queues the specified message to be sent to the remote
func (p *Peer) QueueMessage(m *transport.Message) {
	go func() { p.sendQueue <- m }()
}

// QueueReply queues the specified message to be sent to the remote and appends the desired rpcid
func (p *Peer) QueueReply(rpcID uint64, m *transport.Message) {
	go func() {
		m.RpcId = rpcID
		m.IsReply = true
		p.sendQueue <- m
	}()
}

// Request will ask a remote peer and wait for the response
func (p *Peer) Request(timeout time.Duration, m *transport.Message) (*transport.Message, error) {
	// Create and assign an ID
	current := p.rcpID.Inc()
	m.RpcId = current
	m.IsRequest = true

	// Make a channel to receive the message
	receiveChan := make(chan *transport.Message)

	// Store this so the reply message gets written to it
	p.requestsMux.Lock()
	p.requests[current] = receiveChan
	p.requestsMux.Unlock()

	// Cleanup when we're done
	defer func() {
		p.requestsMux.Lock()
		delete(p.requests, current)
		p.requestsMux.Unlock()
		close(receiveChan)
	}()

	// logger.Info().Field("type", m.Type).Field("is_reply", m.IsReply).Field("remote", p.remote.String()).Log("sent message")

	// Send the message to the remote
	p.QueueMessage(m)

	// Wait for a response or timeout
	select {
	case res := <-receiveChan:
		return res, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("request timed out with timout: %s, request type: %s, request channel: %d, remote: %s", timeout.String(), m.Type, m.RpcId, p.remote.String())
	}
}

// IncomingMessages registers a new listen channel and returns it
func (p *Peer) IncomingMessages() chan *transport.Message {
	return p.receiveChan
}

// BlockUntilDisconnected blocks until the remote is disconnected
func (p *Peer) BlockUntilDisconnected() {
	<-p.session.CloseChan()
}

// CreateClient takes an outgoing connection and creates a client session from it
func (p *Peer) CreateClient(conn net.Conn) error {
	// Setup client side of yamux
	c := yamux.DefaultConfig()
	c.LogOutput = &logWriter{}
	session, err := yamux.Client(conn, c)
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
	c := yamux.DefaultConfig()
	c.LogOutput = &logWriter{}
	session, err := yamux.Server(conn, c)
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
	close(p.receiveChan)
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

func (p *Peer) sendMessage(m *transport.Message) {
	stream, err := p.session.OpenStream()
	defer func() {
		if stream != nil {
			stream.Close()
		}
	}()

	if err != nil {
		logger.Warn().Field("err", err.Error()).Log("peer: error opening connection")
		return
	}

	messageBytes, err := m.Marshal()
	if err != nil {
		logger.Warn().Field("err", err.Error()).Log("peer: error marshalling message")
		return
	}

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
			logger.Debug().Field("err", err.Error()).Log("Error reading message header")
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
		n, err := stream.Read(buffer[numBytesRead:]) // Make sure we don't write over anything
		if err != nil {
			logger.Debug().Field("err", err).Log("Error reading message")
			return
		}
		numBytesRead += n
	}
	// Unmarshal the message
	m := &transport.Message{}
	err = proto.Unmarshal(buffer, m)
	if err != nil {
		logger.Debug().Field("remote_peer", stream.RemoteAddr().String()).Log("peer: could not decode incoming message")
		return
	}

	if utils.LegionAddressFromString(m.Sender).Host != utils.LegionAddressFromString(p.session.RemoteAddr().String()).Host {
		logger.Debug().Field("reported_address", m.Sender).Field("remote_address", stream.RemoteAddr().String()).Log("peer: mismatched reported address and actual remote, disconnecting...")
		p.Close()
		return
	}

	// If this is an RPC response pass it along  to the correct receive channel, if not send it to the
	// regular message receive channels
	if !p.session.IsClosed() {
		if m.IsReply {
			p.requestsMux.Lock()
			respChan, exists := p.requests[m.RpcId]
			p.requestsMux.Unlock()
			if exists {
				go func(c chan *transport.Message) { c <- m }(respChan)
			} else {
				logger.Warn().Field("type", m.Type).Field("local", p.session.LocalAddr().String()).Field("channel_id", m.RpcId).Field("remote", p.remote.String()).Log("Got response to nonexistant RPC channel")
			}
		} else {
			go func(c chan *transport.Message) { c <- m }(p.receiveChan)
		}
	}
}
