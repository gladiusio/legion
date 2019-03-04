package ethpool

import (
	"crypto/ecdsa"
	"errors"
	"sync"

	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gladiusio/legion/frameworks/ethpool/protobuf"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/utils"

	"github.com/gladiusio/legion/network/transport"
	"github.com/gogo/protobuf/proto"

	"sort"

	log "github.com/gladiusio/legion/logger"

	"time"
)

// IncomingMessage represents an incoming message after parsing
type IncomingMessage struct {
	Sender *protobuf.ID
	Body   proto.Message
	Type   string
}

// New returns a Framework that uses the specified function to check if an address is valid, if
// nil all addresses will be considered valid
func New(addressValidator func(common.Address) bool, privKey *ecdsa.PrivateKey) *Framework {
	return &Framework{
		key:              privKey,
		addressValidator: addressValidator,
		messageChan:      make(chan *IncomingMessage),
		idMap:            &sync.Map{},
	}
}

// Framework is a framework for interacting with other peers using ethereum signatures and a kademlia style DHT,
// and only accepting messages from peers that are specified as valid.
type Framework struct {
	// Inherit methods we don't use
	network.GenericFramework

	// Used to check if an address is acceptable
	addressValidator func(common.Address) bool

	l *network.Legion

	// Our Kademlia DHT
	router *RoutingTable

	// Private key to sign messages
	key *ecdsa.PrivateKey

	// The ID of this node
	self *ID

	messageChan chan *IncomingMessage

	// Keep track of ID's and network addresses in an efficient way
	idMap *sync.Map
}

// Assert the type is correct
var _ network.Framework = (*Framework)(nil)

// Configure is used to set up our keystore, and block until we are ready to send/receive messages
func (f *Framework) Configure(l *network.Legion) error {
	f.l = l
	id := &ID{
		EthAddress:     crypto.PubkeyToAddress(f.key.PublicKey).Bytes(),
		NetworkAddress: l.Me().String(),
	}
	f.self = id

	f.router = CreateRoutingTable(*id)

	return nil
}

// ValidateMessage is called before any message is passed to the framework NewMessage()
func (f *Framework) ValidateMessage(ctx *network.MessageContext) bool {
	sm := &protobuf.SignedDHTMessage{}
	err := sm.Unmarshal(ctx.Message.Body)
	if err != nil {
		return false
	}

	// Hash message
	hash := crypto.Keccak256(sm.DhtMessage)

	// Get the public key and address
	pubKey, err := crypto.SigToPub(hash, sm.Signature)
	if err != nil {
		return false
	}
	addr := crypto.PubkeyToAddress(*pubKey)

	// Verify the signature
	if !crypto.VerifySignature(crypto.CompressPubkey(pubKey), hash, sm.Signature[:64]) {
		return false
	}

	m := &protobuf.DHTMessage{}
	err = m.Unmarshal(sm.DhtMessage)
	if err != nil {
		return false
	}

	// Make sure there isn't a nil sender
	if m.GetSender() == nil {
		return false
	}

	// Make sure the sender matches the DHT message
	if !bytes.Equal(m.GetSender().EthAddress, addr.Bytes()) {
		return false
	}

	// Validate that the sender network address matches what is signed
	if ctx.Sender.String() != m.GetSender().NetworkAddress {
		// Disconnect the peer
		ctx.Legion.DeletePeer(ctx.Sender)
		return false
	}

	// Finally check to see if the address is part of the pool
	return f.addressValidator(addr)
}

// Bootstrap will ping any connected nodes with a DHT message
func (f *Framework) Bootstrap() {
	m, err := f.makeLegionSignedMessage("dht.ping", []byte{})
	if err != nil {
		return
	}
	f.l.Broadcast(m)
}

// SendMessage will send a signed version of the message to specified recipient
// it will error if the recipient can't be connected to or found
func (f *Framework) SendMessage(recipient common.Address, messageType string, body proto.Message) error {
	toFind := ID{EthAddress: recipient.Bytes()}
	peers := f.router.FindClosestPeers(toFind, 1)

	if len(peers) != 1 {
		return errors.New("ethpool: could not find peer in routing table, try finding it first")
	}

	if !bytes.Equal(peers[0].EthAddress, toFind.EthAddress) {
		return errors.New("ethpool: could not find peer in routing table, try finding it first")
	}

	bodyBytes, err := proto.Marshal(body)
	if err != nil {
		return errors.New("ethpool: could not marshal message body")
	}

	m, err := f.makeLegionSignedMessage(messageType, bodyBytes)
	if err != nil {
		return errors.New("ethpool: could not make legion signed message")
	}

	la := utils.LegionAddressFromString(peers[0].NetworkAddress)

	f.l.Broadcast(m, la)

	return nil
}

// RecieveMessageChan returns a channel that receives messages
func (f *Framework) RecieveMessageChan() chan *IncomingMessage {
	return f.messageChan
}

// Address returns the ethereum address registered with the framework
func (f *Framework) Address() common.Address {
	return crypto.PubkeyToAddress(f.key.PublicKey)
}

// NewMessage is called when a message is received by the network
func (f *Framework) NewMessage(ctx *network.MessageContext) {
	sm := &protobuf.SignedDHTMessage{}
	err := sm.Unmarshal(ctx.Message.Body)
	if err != nil {
		return
	}

	// Check to see if the signed Ethereum Address matches sender
	dhtMessage := &protobuf.DHTMessage{}
	err = dhtMessage.Unmarshal(sm.DhtMessage)
	if err != nil {
		return
	}

	// Update our router and ID map on all messages
	f.router.Update(ID(*dhtMessage.Sender))
	f.idMap.Store(ctx.Sender, ID(*dhtMessage.Sender))

	// Kademlia methods
	if ctx.Message.Type == "dht.ping" {
		m, err := f.makeLegionSignedMessage("dht.pong", []byte{})
		if err != nil {
			return
		}

		ctx.Reply(m)
	} else if ctx.Message.Type == "dht.pong" {
		f.handlePong()
	} else if ctx.Message.Type == "dht.lookup_request" {
		lookupRequestBytes, err := getDHTMessageBody(ctx.Message.Body)
		if err != nil {
			return
		}

		lookupRequest := &protobuf.LookupRequest{}
		err = lookupRequest.Unmarshal(lookupRequestBytes)
		if err != nil {
			return
		}

		resp := &protobuf.LookupResponse{}

		// Find the closest peers
		for _, peer := range f.router.FindClosestPeers(ID(*lookupRequest.Target), BucketSize) {
			id := protobuf.ID(peer)
			resp.Peers = append(resp.Peers, &id)
		}

		respBytes, err := resp.Marshal()
		if err != nil {
			return
		}

		m, err := f.makeLegionSignedMessage("dht.lookup_response", respBytes)
		if err != nil {
			return
		}
		ctx.Reply(m)

	} else { // Send everything else to the receive channel
		f.messageChan <- &IncomingMessage{Sender: dhtMessage.Sender, Body: dhtMessage, Type: ctx.Message.Type}
	}
}

// PeerDisconnect is called when a peer is deleted
func (f *Framework) PeerDisconnect(ctx *network.PeerContext) {
	id, exists := f.idMap.Load(ctx.Peer.Remote())
	if exists {
		f.router.RemovePeer((id).(ID))
	}
}

func getDHTMessageBody(body []byte) ([]byte, error) {
	sm := &protobuf.SignedDHTMessage{}
	err := sm.Unmarshal(body)
	if err != nil {
		return nil, errors.New("does not look like signed message")
	}

	// Check to see if the signed Ethereum Address matches sender
	m := &protobuf.DHTMessage{}
	err = m.Unmarshal(sm.DhtMessage)
	if err != nil {
		return nil, errors.New("signed message body does not look like dht message")
	}

	return m.Body, nil
}

func (f *Framework) makeLegionSignedMessage(mType string, m []byte) (*transport.Message, error) {
	dhtMessage := &protobuf.DHTMessage{
		Body:   m,
		Sender: (*protobuf.ID)(f.self),
	}

	dhtBytes, err := dhtMessage.Marshal()
	if err != nil {
		return nil, errors.New("ethpool_framework: could not marshal dht message")
	}

	hash := crypto.Keccak256(dhtBytes)

	sig, err := crypto.Sign(hash, f.key)
	if err != nil {
		return nil, errors.New("ethpool_framework: could not sign dht message")
	}

	signedDHTMessage := &protobuf.SignedDHTMessage{
		DhtMessage: dhtBytes,
		Signature:  sig,
	}

	signedBytes, err := signedDHTMessage.Marshal()
	if err != nil {
		return nil, errors.New("ethpool_framework: could not marshal signed dht message")
	}

	return f.l.NewMessage(mType, signedBytes), nil
}

func (f *Framework) handlePong() {
	// Find peers from all the closest remotes
	peers, err := f.findPeers(*f.self, BucketSize)
	if err != nil {
		return
	}

	for _, p := range peers {
		f.router.Update(*p)
	}

}

// Find the peers closest to the ethereum address given
func (f *Framework) findPeers(target ID, count int) ([]*ID, error) {
	// Get our currently connected peers and ask them for the closest to the target
	wg, mux := &sync.WaitGroup{}, &sync.Mutex{}
	peers := make([]*ID, 0)

	for _, peerID := range f.router.FindClosestPeers(target, count) {
		wg.Add(1)

		// Create the request
		tID := protobuf.ID(target)
		lookupRequest := &protobuf.LookupRequest{Target: &tID}
		b, err := lookupRequest.Marshal()
		if err != nil {
			return nil, err
		}
		m, err := f.makeLegionSignedMessage("dht.lookup_request", b)
		if err != nil {
			return nil, err
		}

		// Preform the lookup
		go func(p ID) {
			defer wg.Done()

			incoming, err := f.l.Request(m, time.Second, utils.LegionAddressFromString(p.NetworkAddress))
			if err != nil {
				log.Warn().Field("err", err.Error()).Field("peer", p.EthereumAddress()).Log("Request for lookup was not returned")
				return
			}

			responseBytes, err := getDHTMessageBody(incoming.Body)
			if err != nil {
				return
			}

			lookupResponse := &protobuf.LookupResponse{}
			err = lookupResponse.Unmarshal(responseBytes)
			if err != nil {
				return
			}
			// Convert the type
			toAppend := make([]*ID, len(lookupResponse.GetPeers()))
			for i, id := range lookupResponse.GetPeers() {
				toAppend[i] = (*ID)(id)
			}

			mux.Lock()
			peers = append(peers, toAppend...)
			mux.Unlock()
		}(peerID)
	}

	wg.Wait()

	// Sort resulting peers by XOR distance.
	sort.Slice(peers, func(i, j int) bool {
		left := peers[i].Xor(target)
		right := peers[j].Xor(target)
		return left.Less(right)
	})

	// Cut off list of results to only have the routing table focus on the
	// #dht.BucketSize closest peers to the current node.
	if len(peers) > BucketSize {
		peers = peers[:BucketSize]
	}

	return peers, nil

}
