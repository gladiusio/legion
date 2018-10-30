package network

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"net"

	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
	"github.com/hashicorp/yamux"
)

// NewPeer returns a new peer from the given remote. It also
// sets up the reading and writing channels
func NewPeer(remote *utils.LegionAddress) *Peer {
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
	remote *utils.LegionAddress

	// The internal channel we write to to send a new message
	// to the remote
	sendQueue chan *message.Message

	// The channel of incoming messages
	receiveChans [](chan *message.Message)

	session *yamux.Session
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

// CreateSession takes an incoming connection and creates a session from
func (p *Peer) CreateSession(conn net.Conn) error {
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

func (p *Peer) sendMessage(m *message.Message) {
	stream, err := p.session.Open()
	if err != nil {
		// TODO: Log error
		return
	}

	messageBytes, err := json.Marshal(m)
	if err != nil {
		// TODO: Log error
		return
	}

	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(len(messageBytes)))

	buffer = append(buffer, messageBytes...)

	bw := bufio.NewWriter(stream)
	_, err = bw.Write(buffer)
	if err != nil {
		// TODO: Log error
		return
	}
	bw.Flush()
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
	var err error
	buffer := make([]byte, 4)

	// Read the message size header
	n, numBytesRead := 0, 0
	for numBytesRead < 4 {
		n, err = conn.Read(buffer)
		if err != nil {
			// TODO: Log error
			return
		}

		numBytesRead += n
	}
	// Convert it into an int
	size := binary.BigEndian.Uint32(buffer)
	if size == 0 {
		return
	}

	// Allocate the message size
	buffer = make([]byte, size)

	// Read into the buffer
	n, numBytesRead = 0, 0
	for numBytesRead < int(size) {
		n, err = conn.Read(buffer[numBytesRead:]) // Make sure we don't write over anthing
		if err != nil {
			// TODO: Log error
			return
		}
		numBytesRead += n
	}

	// Unmarshal the message
	m := &message.Message{}
	err = json.Unmarshal(buffer, m)
	if err != nil {
		// TODO: Log error
		return
	}

	// Send off our message into the recieve chans
	for _, rchan := range p.receiveChans {
		go func(c chan *message.Message) { c <- m }(rchan)
	}

	// Close this message stream
	conn.Close()
}
