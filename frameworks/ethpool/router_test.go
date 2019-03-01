package ethpool

import (
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func makeID() *ID {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	id := &ID{
		EthAddress:     crypto.PubkeyToAddress(privKey.PublicKey).Bytes(),
		NetworkAddress: "localhost:6000",
	}

	return id
}

func TestDuplicateEntry(t *testing.T) {
	id := makeID()
	router := CreateRoutingTable(*id)

	peers := router.FindClosestPeers(*id, 4)
	if len(peers) != 0 {
		t.Error("incorrect number of peers")
	}

	id2 := makeID()
	router.Update(*id2)

	peers = router.FindClosestPeers(*id, 4)
	if len(peers) != 1 {
		t.Error("incorrect number of peers")
	}

	router.Update(*id2)

	peers = router.FindClosestPeers(*id, 4)
	if len(peers) != 1 {
		t.Error("incorrect number of peers")
	}

	router.Update(*id)

	peers = router.FindClosestPeers(*id, 4)
	if len(peers) != 1 {
		t.Error("incorrect number of peers")
	}
}
