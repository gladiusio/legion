package ethdht

import (
	"github.com/gladiusio/legion/network"
)

// Plugin is a plugin that uses Ethereum signing/cryptography and protobufs to build a
// kademlia-like DHT for peer messaging and discovery
type Plugin struct {
	// Inherit the methods we don't need so we still meet the interface
	network.GenericPlugin

	l *network.Legion
}

// Compile time assertion that the plugin meets the interface requirements
var _ network.PluginInterface = (*Plugin)(nil)

// Startup is called when the network starts up
func (p *Plugin) Startup(ctx *network.NetworkContext) {
	p.l = ctx.Legion
}

// Bootstrap con
func (p *Plugin) Bootstrap() {
	go func() {
		if p.l != nil {
			p.l.Broadcast(p.l.NewMessage("new_peer_intro", []byte{}))
		}
	}()
}

// NewMessage is called when a new message is received by the network
func (p *Plugin) NewMessage(ctx *network.MessageContext) {
	// Validate the message is of DHT type and warn if not

	// Update the routing table

	// Respond to specific queries
}
