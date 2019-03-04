package clients

import (
	"fmt"
	"time"
)

// Receiver makes a receiver
func Receiver(port uint16) {
	f, _ := makeLegion(port)

	fmt.Printf("Ethereum address: %s\n", f.Address().Hex())

	rChan := f.RecieveMessageChan()

	fmt.Println("Waiting for messages...")
	// Block until start message
	for m := range rChan {
		if m.Type == "start" {
			break
		}
	}

	fmt.Println("Got first message, benchmarking...")
	count := 0
	start := time.Now()
	var taken time.Duration
	for m := range rChan {
		if m.Type == "stop" {
			taken = time.Since(start)
			break
		}
		count++
	}
	fmt.Printf("Received %d messages over %s at a rate of %f messages/second\n", count, taken.String(), float64(count)/taken.Seconds())
}
