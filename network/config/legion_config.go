package config

import (
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
)

// LegionConfig is a config object for the legion network
type LegionConfig struct {
	BindAddress      utils.LegionAddress
	MessageValidator func(*message.Message) bool
}
