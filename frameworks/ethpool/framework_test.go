package ethpool

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gladiusio/legion/frameworks/ethpool/protobuf"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/utils"

	"sync"
	"testing"
	"time"
)

func makeConfig(port uint16) *config.LegionConfig {
	return &config.LegionConfig{
		BindAddress:      utils.NewLegionAddress("localhost", port),
		AdvertiseAddress: utils.NewLegionAddress("localhost", port),
	}
}

func newFrameworkGroup(n int) *frameworkGroup {
	l := &frameworkGroup{frameworks: make([]*Framework, n)}
	l.makeFrameworks(n)
	return l
}

type frameworkGroup struct {
	frameworks []*Framework
	legions    []*network.Legion
}

func (lg *frameworkGroup) makeFrameworks(n int) {
	frameworks := make([]*Framework, 0, n)
	legions := make([]*network.Legion, 0, n)

	for i := 0; i < n; i++ {
		privKey, err := crypto.GenerateKey()
		if err != nil {
			panic(err)
		}
		f := New(func(string) bool { return true }, privKey)
		l := network.NewLegion(makeConfig(7000+uint16(i)), f)
		go func() {
			err := l.Listen()
			if err != nil {
				panic(err)
			}
		}()

		frameworks = append(frameworks, f)
		legions = append(legions, l)
	}

	lg.frameworks = frameworks
	lg.legions = legions
}

func (lg *frameworkGroup) connect() {
	for i := 1; i < len(lg.frameworks); i++ {
		err := lg.legions[0].AddPeer(lg.legions[i].Me())
		if err != nil {
			panic(err)
		}
	}
}

func (lg *frameworkGroup) waitUntilStarted() {
	var wg sync.WaitGroup
	for _, leg := range lg.legions {
		wg.Add(1)
		go func(l *network.Legion) {
			l.Started()
			wg.Done()
		}(leg)
	}

	wg.Wait()
}

func (lg *frameworkGroup) stop() {
	for _, leg := range lg.legions {
		err := leg.Stop()
		if err != nil {
			panic(err)
		}
	}
}

func TestBootstrap(t *testing.T) {
	fg := newFrameworkGroup(3)
	fg.waitUntilStarted()
	fg.connect()

	for _, f := range fg.frameworks {
		f.Bootstrap()
		time.Sleep(15 * time.Millisecond)
	}

	for _, f := range fg.frameworks {
		peers := f.router.GetPeers()
		if len(peers) != 2 {
			t.Errorf("Incorrect number of peers, should have been 2, was %d", len(peers))
		}
	}

	fg.stop()
}

func TestMessaging(t *testing.T) {
	fg := newFrameworkGroup(3)
	fg.waitUntilStarted()
	fg.connect()

	for _, f := range fg.frameworks {
		f.Bootstrap()
		time.Sleep(10 * time.Millisecond)
	}

	toSend := fg.frameworks[0].self
	fg.frameworks[1].SendMessage(toSend.EthereumAddress(), "testing", &protobuf.Empty{})

	receiveChan := fg.frameworks[0].RecieveMessageChan()
	noReceiveChan1 := fg.frameworks[1].RecieveMessageChan()
	noReceiveChan2 := fg.frameworks[2].RecieveMessageChan()

	received := false
L:
	for {
		select {
		case res := <-receiveChan:
			if res.Type != "testing" {
				t.Errorf("Received bad message type: %s", res.Type)
			}
			received = true
		case <-noReceiveChan1:
			t.Error("Non addressed peer received message")
		case <-noReceiveChan2:
			t.Error("Non addressed peer received message")
		case <-time.After(2 * time.Second):
			break L
		}
	}

	if !received {
		t.Error("Recipient never got message")
	}

	fg.stop()

}
