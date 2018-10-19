/*
Package legion is a simple top level package to allow a nice import syntax
*/
package legion

import (
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/utils"
)

// New returns a new Legion object which contains most of the function needed to work with the network.
func New(c *config.LegionConfig) *network.Legion {
	return network.NewLegion(c)
}

// SimpleConfig returns a safe config with only the bind address and port specified
func SimpleConfig(bindAddress string, port uint16) *config.LegionConfig {
	return &config.LegionConfig{BindAddress: utils.NewLegionAddress(bindAddress, port)}
}
