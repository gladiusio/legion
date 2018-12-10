package main

import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gladiusio/legion"
	"github.com/gladiusio/legion/examples/chat/plugin"
	"github.com/gladiusio/legion/plugins/simpledisc"
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
	disc := new(simpledisc.Plugin)
	l.RegisterPlugin(disc) // Add the basic discovery plugin
	go l.Listen()
	l.Started()

	if *remote != "" {
		err := l.PromotePeer(utils.LegionAddressFromString(*remote))
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Reach out to the peers we just connected to
	disc.Bootstrap()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadBytes('\n')
		l.Broadcast(l.NewMessage("chat_message", text))
	}
}
