package plugin

import (
	"fmt"

	"github.com/gladiusio/legion/network"
)

type MessagePlugin struct {
	network.GenericPlugin
	specialData string
}

func (m *MessagePlugin) NewMessage(ctx *MessageContext) {
	fmt.Println(mspecialData)
}
