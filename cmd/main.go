package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gladiusio/gladius-p2p/p2p"
)

func main() {
	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	node := &p2p.Node{}
	node.SetIncommingMessageHandler(handleConnection)

	if *target != "" {
		address := strings.Split(*target, ":")[0]
		port := strings.Split(*target, ":")[1]

		peer := &p2p.Peer{Address: address, Port: port}
		node.AddPeer(peer)
	}
	if *target != ":" {
		i := strconv.Itoa(*listenF)
		node.SetUpHost(i, "localhost")
	}
	node.Start()

	fmt.Println("Finished starting node")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		node.SendMessage(scanner.Text())
	}

	if scanner.Err() != nil {
		// handle error.
	}
}

func handleConnection(rw *bufio.ReadWriter) {
	log.Println("Got a new stream")

	go readData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(str)
	}
}

func writeData(rw *bufio.ReadWriter) {
}
