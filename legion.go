/*
Package legion is a simple top level package to allow a nice import syntax
*/
package legion

import (
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
)

// New returns a new Legion object which contains most of the function needed to work with the network.
func New(c *config.LegionConfig) *network.Legion {
	return network.NewLegion(c)
}

// DefaultConfig returns a config with only the bind address and port specified,
// and all messages are considered valid
func DefaultConfig(bindAddress string, port uint16) *config.LegionConfig {
	return &config.LegionConfig{
		BindAddress:      utils.NewLegionAddress(bindAddress, port),
		AdvertiseAddress: utils.NewLegionAddress(bindAddress, port),
		MessageValidator: func(m *message.Message) bool { return true },
	}
}
