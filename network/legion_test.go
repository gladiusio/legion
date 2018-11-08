package network

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/events"
	"github.com/gladiusio/legion/network/message"
	"github.com/gladiusio/legion/utils"
)

func makeConfig(port uint16) *config.LegionConfig {
	return &config.LegionConfig{
		BindAddress:      utils.NewLegionAddress("localhost", port),
		MessageValidator: func(m *message.Message) bool { return true },
	}
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
	callback func(ctx *MessageContext)
}

func (m *MessagePlugin) NewMessage(ctx *MessageContext) {
	m.callback(ctx)
}

func TestFireMessageEvent(t *testing.T) {
	l := NewLegion(makeConfig(6000))
	failed := true
	p := &MessagePlugin{callback: func(ctx *MessageContext) {
		failed = false
	}}
	l.RegisterPlugin(p)

	l.FireMessageEvent(events.NewMessageEvent, &message.Message{})

	time.Sleep(50 * time.Millisecond)

	if failed {
		t.Error("message event never fired")
	}
}

func newLegionGroup(n int) *legionGroup {
	l := &legionGroup{legions: make([]*Legion, n)}
	l.makeLegions(n)
	return l
}

type legionGroup struct {
	legions []*Legion
}

func (lg *legionGroup) makeLegions(n int) {
	legions := make([]*Legion, 0, n)
	for i := 0; i < n; i++ {
		l := NewLegion(makeConfig(6000 + uint16(i)))
		go func() {
			err := l.Listen()
			if err != nil {
				panic(err)
			}
		}()
		legions = append(legions, l)
	}

	lg.legions = legions
}

func (lg *legionGroup) connect() {
	for i := 1; i < len(lg.legions); i++ {
		err := lg.legions[0].AddPeer(lg.legions[i].config.BindAddress)
		if err != nil {
			panic(err)
		}
	}
}

func (lg *legionGroup) waitUntilStarted() {
	var wg sync.WaitGroup
	for _, leg := range lg.legions {
		wg.Add(1)
		go func(l *Legion) {
			l.Started()
			wg.Done()
		}(leg)
	}

	wg.Wait()
}

func (lg *legionGroup) stop() {
	for _, leg := range lg.legions {
		err := leg.Stop()
		if err != nil {
			panic(err)
		}
	}
}

func TestPeerConnection(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	time.Sleep(100 * time.Millisecond)

	peerCount := 0
	lg.legions[0].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	peerCount = 0
	lg.legions[1].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}

func TestPeerConnectionWhenMessageRecieved(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	time.Sleep(100 * time.Millisecond)

	peerCount := 0
	lg.legions[0].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	// Peer 1 sends introduction to peer 2
	lg.legions[0].Broadcast(message.New(lg.legions[0].config.BindAddress, "test", []byte{}), lg.legions[1].config.BindAddress)

	time.Sleep(100 * time.Millisecond)

	peerCount = 0
	lg.legions[1].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}

func TestPromotePeer(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	time.Sleep(100 * time.Millisecond)

	lg.legions[0].PromotePeer(lg.legions[1].config.BindAddress)

	peerCount := 0
	lg.legions[0].promotedPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("promoted number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	peerCount = 0
	lg.legions[0].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("number of total peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}

func TestBroadcast(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	failed := true
	p := &MessagePlugin{callback: func(ctx *MessageContext) {
		if ctx.Message.Type() == "test" {
			failed = false
		}
	}}
	lg.legions[1].RegisterPlugin(p)

	lg.legions[0].PromotePeer(lg.legions[1].config.BindAddress)
	time.Sleep(100 * time.Millisecond)

	lg.legions[0].Broadcast(message.New(lg.legions[0].config.BindAddress, "test", []byte{}))

	time.Sleep(200 * time.Millisecond)

	if failed {
		t.Error("peer never received message")
	}
}

func TestBroadcastRandomNGreaterThanPeers(t *testing.T) {
	lg := newLegionGroup(10)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	var count uint64
	p := &MessagePlugin{callback: func(ctx *MessageContext) {
		if ctx.Message.Type() == "test" {
			atomic.AddUint64(&count, 1)
		}
	}}
	for _, leg := range lg.legions[1:] {
		lg.legions[0].PromotePeer(leg.config.BindAddress)
		leg.RegisterPlugin(p)
	}

	lg.legions[0].BroadcastRandom(message.New(lg.legions[0].config.BindAddress, "test", []byte{}), 11)

	time.Sleep(300 * time.Millisecond)

	if count != 9 {
		t.Errorf("random broadcast was not sent to all peers, should have been 9, was: %d", count)
	}
}

func TestBroadcastRandom(t *testing.T) {
	lg := newLegionGroup(10)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	var count uint64
	p := &MessagePlugin{callback: func(ctx *MessageContext) {
		if ctx.Message.Type() == "test" {
			atomic.AddUint64(&count, 1)
		}
	}}
	for _, leg := range lg.legions[1:] {
		lg.legions[0].PromotePeer(leg.config.BindAddress)
		leg.RegisterPlugin(p)
	}

	lg.legions[0].BroadcastRandom(message.New(lg.legions[0].config.BindAddress, "test", []byte{}), 5)

	time.Sleep(300 * time.Millisecond)

	if count != 5 {
		t.Errorf("random broadcast was not sent to all peers, should have been 5, was: %d", count)
	}
}

func TestSelfDial(t *testing.T) {

}

func TestSingleConnectionOpened(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	time.Sleep(100 * time.Millisecond)

	// Peer 1 sends introduction to peer 2
	lg.legions[0].Broadcast(message.New(lg.legions[0].config.BindAddress, "test", []byte{}), lg.legions[1].config.BindAddress)

	time.Sleep(300 * time.Millisecond)

	peerCount := 0
	lg.legions[1].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect after intro message, there should have been 1, there were: %d", peerCount)
	}

	// Peer 2 sends message to peer 1
	lg.legions[1].Broadcast(message.New(lg.legions[1].config.BindAddress, "test", []byte{}), lg.legions[0].config.BindAddress)

	time.Sleep(100 * time.Millisecond)

	peerCount = 0
	lg.legions[0].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	peerCount = 0
	lg.legions[1].allPeers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}
