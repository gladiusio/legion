package clients

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gladiusio/legion/frameworks/ethpool"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/network/config"
	"github.com/gladiusio/legion/utils"
)

func makeLegion(port uint16) (*ethpool.Framework, *network.Legion) {
	conf := &config.LegionConfig{
		BindAddress:      utils.NewLegionAddress("localhost", port),
		AdvertiseAddress: utils.NewLegionAddress("localhost", port),
	}

	privKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	f := ethpool.New(func(common.Address) bool { return true }, privKey)
	l := network.NewLegion(conf, f)
	go func() {
		err := l.Listen()
		if err != nil {
			panic(err)
		}
	}()

	return f, l
}
