package statenode

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/gladiusio/gladius-p2p/p2p"
)

// New - Returns a new StateNode
func New(target string, listenAddress string, listenPort string) StateNode {
	node := &p2p.Node{}
	node.SetIncommingMessageHandler(handleConnection)

	if target != "" {
		address := strings.Split(target, ":")[0]
		port := strings.Split(target, ":")[1]

		peer := &p2p.Peer{Address: address, Port: port}
		node.AddPeer(peer)
	}
	if target != ":" {
		node.SetUpHost(listenPort, listenAddress)
	}
	node.Start()

	node.SendMessage("Announcing Me!")

	return StateNode{}
}

// StateNode - Node with state and identity
type StateNode struct {
	node *p2p.Node
}

func (sn *StateNode) PushState() {

}

func (sn StateNode) StateOf( /*Need an address here*/ ) {

}

func (sn *StateNode) RegisterIdentity() {

}

func handleConnection(rw *bufio.ReadWriter) {
	log.Println("Got a new message input stream")

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Print(err)
			break
		}

		fmt.Println(str)
	}
}
