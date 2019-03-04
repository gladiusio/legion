package clients

import (
	"crypto/rand"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gladiusio/legion/frameworks/ethpool/protobuf"
	"github.com/gladiusio/legion/utils"
)

// Sender makes a sender
func Sender(receiverEth, remoteAddress string, listenPort uint16, numOfMessages int) {
	if receiverEth == "" {
		log.Fatal("Need to set -remote_eth to the correct address")
	}
	body := make([]byte, 100)
	rand.Read(body)

	var i uint16
	for i = 0; i < 40; i++ {
		go func(offset uint16) {
			f, l := makeLegion(listenPort + offset)
			l.AddPeer(utils.LegionAddressFromString(remoteAddress))
			f.Bootstrap()

			time.Sleep(1 * time.Second)

			addr := common.HexToAddress(receiverEth)
			f.SendMessage(addr, "start", &protobuf.Empty{})

			for i := 0; i < numOfMessages; i++ {
				f.SendMessage(addr, "bench", &protobuf.Empty{})
			}

			f.SendMessage(addr, "stop", &protobuf.DHTMessage{
				Body: body,
			})
		}(i)
	}

	select {}
}
