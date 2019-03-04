package main

import (
	"flag"
	"log"

	"github.com/gladiusio/legion/benchmarks/clients"
)

func main() {
	isSender := flag.Bool("sender", false, "Set up the benchmark as a sender")
	isReceiver := flag.Bool("receiver", false, "Set up the benchmark as a receiver")
	port := flag.Int("port", 6000, "Bind port for the sender")
	numOfMessages := flag.Int("messages", 10000, "The number of messages to send to the remote")
	remote := flag.String("remote", "localhost:6000", "Network address of the receiver")
	receiverEth := flag.String("remote_eth", "", "Ethereum address of the receiver")

	flag.Parse()

	if *isReceiver && *isSender {
		log.Fatal("Cannot be sender and receiver")
	}

	if !(*isReceiver || *isSender) {
		log.Fatal("Must specify receiver or sender flag")
	}

	if *isReceiver {
		clients.Receiver(uint16(*port))
	} else {
		clients.Sender(*receiverEth, *remote, uint16(*port), *numOfMessages)
	}
}
