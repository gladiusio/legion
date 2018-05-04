package p2p

import (
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
}

// Serialize - Get a serialized peer like address:port
func (p *Peer) Serialize() (serialized string) {
	return fmt.Sprintf("%s:%s", p.Address, p.Port)
}

// Node - Struct representing a node in the p2p network.
type Node struct {
	allPeers []Peer
	seedNode *Peer
	maxPeers int
	handler  func(net.Conn)

	isHost      bool
	hostAddress string
	hostPort    string
}

// SetStreamHandler - Set the function to handle the stream on data, in or out.
func (n *Node) SetStreamHandler(handler func(net.Conn)) {
	n.handler = handler
}

// SetSeedNode - Set the seed node that is used to detect other nodes in the network
func (n *Node) SetSeedNode(seed *Peer) {
	n.seedNode = seed
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

// TODO: Auth
func (n Node) connectToHost(p *Peer) {
	conn, _ := net.Dial("tcp", p.Serialize())
	go n.handler(conn)
}

func startHostListener(n *Node) {
	// Start listening to port 8888 for TCP connection
	listener, err := net.Listen("tcp", n.makeHostString())
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		listener.Close()
		fmt.Println("Listener closed")
	}()

	// Get net.TCPConn object
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
	}

	// Use their handler
	go n.handler(conn)

}

// Start - Starts the node's discovery, and opens a stream to a subset of peerList.
func (n *Node) Start() {
	if n.seedNode == nil && len(n.allPeers) > 0 { // No seed node provided
		rand.Seed(time.Now().Unix())                           // initialize global pseudo random generator
		n.SetSeedNode(&n.allPeers[rand.Intn(len(n.allPeers))]) // Pick a random peer to get seed node
	} else {
		panic(errors.New("No seed node could be found"))
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
