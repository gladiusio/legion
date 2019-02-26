package network

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/network/events"
	"github.com/gladiusio/legion/network/transport"
	"github.com/gladiusio/legion/utils"
)

func makeConfig(port uint16) *config.LegionConfig {
	return &config.LegionConfig{
		BindAddress:      utils.NewLegionAddress("localhost", port),
		AdvertiseAddress: utils.NewLegionAddress("localhost", port),
	}
}

func TestLegionCreation(t *testing.T) {
	l := NewLegion(makeConfig(6000), nil)

	if l.peers == nil {
		t.Error("peers was not initialized")
	}

	if l.framework == nil {
		t.Error("framework was not initialized")
	}
}

func TestFrameowkr(t *testing.T) {
	l := NewLegion(makeConfig(6000), new(GenericFramework))

	if l.framework == nil {
		t.Errorf("framework not added")
	}
}

type MessageFramework struct {
	GenericFramework
	callback func(ctx *MessageContext)
}

func (m *MessageFramework) NewMessage(ctx *MessageContext) {
	m.callback(ctx)
}

func TestFireMessageEvent(t *testing.T) {
	failed := true

	f := &MessageFramework{callback: func(ctx *MessageContext) {
		failed = false
	}}
	l := NewLegion(makeConfig(6000), f)

	l.FireMessageEvent(events.NewMessageEvent, &transport.Message{})

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
		l := NewLegion(makeConfig(6000+uint16(i)), nil)
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
	lg.legions[0].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	// Send introduction
	lg.legions[0].Broadcast(lg.legions[0].NewMessage("test", []byte{}))

	time.Sleep(100 * time.Millisecond)

	peerCount = 0
	lg.legions[1].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
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
	lg.legions[0].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	// Peer 1 sends introduction to peer 2
	lg.legions[0].Broadcast(lg.legions[0].NewMessage("", []byte{}), lg.legions[1].config.BindAddress)

	time.Sleep(100 * time.Millisecond)

	peerCount = 0
	lg.legions[1].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}

func TestBroadcast(t *testing.T) {
	failed := true

	f := &MessageFramework{callback: func(ctx *MessageContext) {
		if ctx.Message.GetType() == "test" {
			failed = false
		}
	}}

	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	lg.legions[1].framework = f

	lg.legions[0].Broadcast(lg.legions[0].NewMessage("intro", []byte{}))

	lg.legions[0].Broadcast(lg.legions[0].NewMessage("test", []byte{}))

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
	f := &MessageFramework{callback: func(ctx *MessageContext) {
		if ctx.Message.GetType() == "test" {
			atomic.AddUint64(&count, 1)
		}
	}}

	for _, leg := range lg.legions[1:] {
		leg.framework = f
	}

	time.Sleep(100 * time.Millisecond)

	lg.legions[0].BroadcastRandom(lg.legions[0].NewMessage("test", []byte{}), 11)

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
	f := &MessageFramework{callback: func(ctx *MessageContext) {
		if ctx.Message.GetType() == "test" {
			atomic.AddUint64(&count, 1)
		}
	}}
	for _, leg := range lg.legions[1:] {
		leg.framework = f
	}

	time.Sleep(100 * time.Millisecond)

	lg.legions[0].BroadcastRandom(lg.legions[0].NewMessage("test", []byte{}), 5)

	time.Sleep(300 * time.Millisecond)

	if count != 5 {
		t.Errorf("random broadcast was not sent to all peers, should have been 5, was: %d", count)
	}
}

func TestDoFunction(t *testing.T) {
	lg := newLegionGroup(6)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	var count uint64
	f := func(p *Peer) { atomic.AddUint64(&count, 1) }

	lg.legions[0].DoAllPeers(f)
	fmt.Println()

	if count != 5 {
		t.Errorf("function was not called on all peers, should have been 5, was: %d", count)
	}
}

func TestSingleConnectionOpened(t *testing.T) {
	lg := newLegionGroup(2)
	lg.waitUntilStarted()

	lg.connect()
	defer lg.stop()

	time.Sleep(100 * time.Millisecond)

	// Peer 1 sends introduction to peer 2
	lg.legions[0].Broadcast(lg.legions[0].NewMessage("test", []byte{}), lg.legions[1].config.BindAddress)

	time.Sleep(300 * time.Millisecond)

	peerCount := 0
	lg.legions[1].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect after intro message, there should have been 1, there were: %d", peerCount)
	}

	// Peer 2 sends message to peer 1
	lg.legions[1].Broadcast(lg.legions[0].NewMessage("test", []byte{}), lg.legions[0].config.BindAddress)

	time.Sleep(100 * time.Millisecond)

	peerCount = 0
	lg.legions[0].peers.Range(func(key, value interface{}) bool { peerCount++; fmt.Println(key); return true })
	if peerCount != 1 {
		t.Errorf("local number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}

	peerCount = 0
	lg.legions[1].peers.Range(func(key, value interface{}) bool { peerCount++; return true })
	if peerCount != 1 {
		t.Errorf("remote number of peers is incorrect, there should have been 1, there were: %d", peerCount)
	}
}
