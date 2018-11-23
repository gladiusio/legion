package simpledisc

import (
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/utils"
)

// Plugin is a plugin to discover other nodes in a Kademlia like way
type Plugin struct {
	network.GenericPlugin
}

// Compile time assertion that the plugin meets the interface requirements
var _ network.PluginInterface = (*Plugin)(nil)

// Startup is called once the network has started
func (p *Plugin) Startup(ctx *network.NetworkContext) {
	go func() {
		// Block until the network has started
		ctx.Legion.Started()

		// Ask any peers we're connected to for their list of peers
		ctx.Legion.Broadcast(ctx.Legion.NewMessage("new_peer_intro", []byte{}))
	}()
}

// NewMessage is called when a new message is received by the network
func (p *Plugin) NewMessage(ctx *network.MessageContext) {
	mType := ctx.Message.Type()
	if mType == "peer_list" {

	} else if mType == "new_peer" {
		// Add the peer as promoted
		ctx.Legion.PromotePeer(utils.LegionAddressFromString(string(ctx.Message.Body())))
	} else if mType == "new_peer_intro" {
		// Tell all of our peers about the newly connected peer
		peerBytes := []byte(ctx.Message.Sender().String())
		ctx.Legion.Broadcast(ctx.Legion.NewMessage("new_peer", peerBytes))

		// Reply with all of our peers
		ctx.Legion.DoAllPeers(f)
		ctx.Reply(ctx.Legion.NewMessage("peer_list", []byte{}))
	}
}
