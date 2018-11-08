package plugin

import (
	"fmt"

	"github.com/gladiusio/legion/network"
	"github.com/logrusorgru/aurora"
)

// ChatPlugin is a plugin to print new messages
type ChatPlugin struct {
	network.GenericPlugin
}

// Startup is called once the network has started
func (p *ChatPlugin) Startup(ctx *network.NetworkContext) {
	ctx.Legion.Started() // Wait until we're listening
	fmt.Printf("[%s] Started chat client, type below and press enter to send messages.\n", aurora.Green("Legion Chat Example"))
}

// NewMessage is called when a new message is recieved by the network
func (p *ChatPlugin) NewMessage(ctx *network.MessageContext) {
	// First we make sure it is the right type
	if ctx.Message.Type() == "chat_message" {
		// Then we print it
		fmt.Printf("[%s] %s", aurora.Blue(ctx.Message.Sender().String()), ctx.Message.Body())
	}
}

// PeerAdded is called when a new peer connects or is added
func (p *ChatPlugin) PeerAdded(ctx *network.PeerContext) {
	ctx.Legion.PromotePeer(ctx.Peer.Remote())
}
