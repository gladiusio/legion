package simpledisc

import (
	"encoding/json"

	"github.com/gladiusio/legion/logger"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/utils"
)

// Plugin is a plugin to discover other nodes in a simple way
type Plugin struct {
	network.GenericPlugin
	l *network.Legion
}

// Compile time assertion that the plugin meets the interface requirements
var _ network.PluginInterface = (*Plugin)(nil)

// Startup is called when the network starts up
func (p *Plugin) Startup(ctx *network.NetworkContext) {
	p.l = ctx.Legion
}

// Bootstrap reaches out to all connected peers and asks them for
// information about other peers
func (p *Plugin) Bootstrap() {
	go func() {
		if p.l != nil {
			p.l.Broadcast(p.l.NewMessage("new_peer_intro", []byte{}))
		}
	}()
}

// NewMessage is called when a new message is received by the network
func (p *Plugin) NewMessage(ctx *network.MessageContext) {
	mType := ctx.Message.Type()
	if mType == "peer_list" {
		peerList := make([]string, 0)
		err := json.Unmarshal(ctx.Message.Body(), &peerList)
		if err != nil {
			return
		}
		connectedPeers := make([]string, 0)
		for _, p := range peerList {
			addr := utils.LegionAddressFromString(p)
			if addr.IsValid() && addr != ctx.Legion.Me() {
				go ctx.Legion.AddPeer(addr)
				connectedPeers = append(connectedPeers, p)
			}
		}
		logger.Info().Field("peers", connectedPeers).Log("Connected to peers")
	} else if mType == "new_peer" {
		// Add the peer as promoted
		addr := utils.LegionAddressFromString(string(ctx.Message.Body()))
		if addr.IsValid() && addr != ctx.Legion.Me() {
			ctx.Legion.PromotePeer(addr)
		}
	} else if mType == "new_peer_intro" {
		logger.Info().Field("peer:", ctx.Message.Sender()).Log("New peer introduction")
		// Tell all of our peers about the newly connected peer
		peerBytes := []byte(ctx.Message.Sender().String())
		ctx.Legion.Broadcast(ctx.Legion.NewMessage("new_peer", peerBytes))

		peerList := make([]string, 0)

		// Function to get a list of all the peers that we know
		f := func(p *network.Peer) { peerList = append(peerList, p.Remote().String()) }

		// Reply with all of our peers
		ctx.Legion.DoAllPeers(f)

		// Encode our list and send it
		b, err := json.Marshal(peerList)
		if err != nil {
			logger.Warn().Log("Error encoding peer list")
			return
		}

		ctx.Reply(ctx.Legion.NewMessage("peer_list", b))
	}
}
