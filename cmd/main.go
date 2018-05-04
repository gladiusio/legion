package main

import (
	"bufio"
	"log"
	"net"

	"github.com/gladiusio/gladius-p2p/p2p"
)

func main() {
	node := &p2p.Node{}
	peer := &p2p.Peer{Address: "localhost", Port: "8001"}

	node.SetSeedNode(peer)
	node.SetStreamHandler(handleConnection)
	node.SetUpHost("8002", "localhost")
	node.Start()

	select {} // hang forever
}

func handleConnection(conn net.Conn) {
	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
}

func writeData(rw *bufio.ReadWriter) {
}
