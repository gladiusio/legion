package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/gladiusio/legion"
	"github.com/gladiusio/legion/examples/chat/plugin"
	"github.com/gladiusio/legion/utils"
)

func main() {
	bindAddress := flag.String("bindaddress", "localhost:6000", "the address to bind to")
	remote := flag.String("remote", "", "the remote address to connect to")

	flag.Parse()

	splitAddress := strings.Split(*bindAddress, ":")

	host := splitAddress[0]
	port, _ := strconv.Atoi(splitAddress[1])

	conf := legion.DefaultConfig(host, uint16(port))
	l := legion.New(conf)
	l.RegisterPlugin(new(plugin.ChatPlugin))
	go l.Listen()
	l.Started()

	if *remote != "" {
		err := l.PromotePeer(utils.LegionAddressFromString(*remote))
		if err != nil {
			panic(err)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadBytes('\n')
		l.Broadcast(l.NewMessage("chat_message", text))
	}
}
