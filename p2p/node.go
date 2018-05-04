package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Peer - Information to connect to a peer
type Peer struct {
	Address string
	Port    string
	rw      *bufio.ReadWriter
}

// Serialize - Get a serialized peer like address:port
func (p *Peer) Serialize() (serialized string) {
	return fmt.Sprintf("%s:%s", p.Address, p.Port)
}

// Node - Struct representing a node in the p2p network.
type Node struct {
	allPeers       []*Peer
	connectedPeers []*Peer
	seedNode       *Peer
	maxPeers       int
	handler        func(*bufio.ReadWriter)

	isHost      bool
	hostAddress string
	hostPort    string
}

// SetIncommingMessageHandler - Set the function to handle the incomming message stream
func (n *Node) SetIncommingMessageHandler(handler func(*bufio.ReadWriter)) {
	n.handler = handler
}

// SendMessage - Sends a message to all peers
func (n *Node) SendMessage(message string) {
	for _, peer := range n.connectedPeers {
		peer.rw.WriteString(fmt.Sprintf("%s\n", message))
		peer.rw.Flush()
	}
}

// SetSeedNode - Set the seed node that is used to detect other nodes in the network
func (n *Node) SetSeedNode(seed *Peer) {
	n.seedNode = seed
}

// AddPeer - Force add a peer by connecting to it.
func (n *Node) AddPeer(peer *Peer) {
	n.connectToHost(peer)
}

// SetUpHost - Sets up a host on the given address and port
func (n *Node) SetUpHost(port string, address string) {
	n.isHost = true
	n.hostPort = port
	n.hostAddress = address
}

func (n Node) makeHostString() string {
	return fmt.Sprintf("%s:%s", n.hostAddress, n.hostPort)
}

func (n *Node) connectToHost(p *Peer) {
	conn, err := net.Dial("tcp", p.Serialize())
	if err != nil {
		panic(err)
	}
	p.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	n.connectedPeers = append(n.connectedPeers, p)
	go func() { n.handler(p.rw); conn.Close() }()
}

func (n *Node) startHostListener() {
	// Start listening to port 8888 for TCP connection
	listener, err := net.Listen("tcp", n.makeHostString())
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			// Get net.TCPConn object
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
			}
			rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
			newPeer := &Peer{Address: conn.RemoteAddr().String(), Port: "", rw: rw}
			n.connectedPeers = append(n.connectedPeers, newPeer)

			go func() { n.handler(rw); conn.Close() }()
		}
	}()

}

// Start - Starts the node's discovery, and opens a stream to a subset of peerList.
func (n *Node) Start() {
	if n.seedNode == nil && len(n.allPeers) > 0 { // No seed node provided
		rand.Seed(time.Now().Unix())                          // initialize global pseudo random generator
		n.SetSeedNode(n.allPeers[rand.Intn(len(n.allPeers))]) // Pick a random peer to get seed node
	} else {
		panic(errors.New("No seed node could be found"))
	}

	if n.isHost {
		n.startHostListener()
	}

	// Find other nodes from the seed node
	n.discover()
}

// SerializePeers - Returns a string of every peer known.
func (n *Node) SerializePeers() (peers []string) {
	toReturn := make([]string, 10)

	// Go through each peer
	for _, peer := range n.allPeers {
		toReturn = append(toReturn, peer.Serialize())
	}

	return toReturn
}

func (n *Node) discover() {
	// FIXME: So inefficient... Use mapping to bool instead of slice
	for _, newPeer := range getNodePeers(n.seedNode) {
		for _, existingPeer := range n.allPeers {
			if newPeer.Address != existingPeer.Address && newPeer.Port != existingPeer.Port {
				n.allPeers = append(n.allPeers, newPeer)
			}
		}
	}
}
