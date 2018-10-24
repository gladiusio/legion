package network

import (
	"testing"
	"time"

	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/events"
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
)

func makeConfig(port uint16) *config.LegionConfig {
	return &config.LegionConfig{BindAddress: utils.NewLegionAddress("localhost", port)}
}

func TestLegionCreation(t *testing.T) {
	l := NewLegion(makeConfig(6000))

	if l.allPeers == nil {
		t.Error("allPeers was not initialized")
	}

	if l.promotedPeers == nil {
		t.Error("promotedPeers was not initialized")
	}

	if l.plugins == nil {
		t.Error("plugin list was not initialized")
	}
}

func TestRegisterPlugin(t *testing.T) {
	l := NewLegion(makeConfig(6000))
	p := new(GenericPlugin)
	l.RegisterPlugin(p)

	if len(l.plugins) != 1 {
		t.Errorf("plugin list length should be 1, was %d", len(l.plugins))
	}
}

type MessagePlugin struct {
	GenericPlugin
	callback func()
}

func (m *MessagePlugin) NewMessage(ctx *MessageContext) {
	m.callback()
}

func TestFireMessageEvent(t *testing.T) {
	l := NewLegion(makeConfig(6000))
	failed := true
	p := &MessagePlugin{callback: func() { failed = false }}
	l.RegisterPlugin(p)

	l.FireMessageEvent(events.NewMessageEvent, &message.Message{})

	time.Sleep(50 * time.Millisecond)

	if failed {
		t.Error("message event never fired")
	}
}

func TestPeerConnection(t *testing.T) {
	l1 := NewLegion(makeConfig(6000))
	l2 := NewLegion(makeConfig(6001))

	go l1.Listen()
	go l2.Listen()

	l1.Started()
	l2.Started()

	l1.AddPeer(l2.config.BindAddress)

	time.Sleep(100 * time.Millisecond)

	peerCount := 0
	l1.allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

}
