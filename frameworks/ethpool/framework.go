package ethpool

import (
	"crypto/ecdsa"
	"errors"
	"sync"

	"bytes"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gladiusio/legion/frameworks/ethpool/protobuf"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/network/transport"
	"github.com/gogo/protobuf/proto"
)

// IncomingMessage represents an incoming message after parsing
type IncomingMessage struct {
	Sender *protobuf.ID
	Body   proto.Message
}

// New returns a Framework that uses the specified function to check if an address is valid, if
// nil all addresses will be considered valid
func New(addressValidator func(string) bool, privKey *ecdsa.PrivateKey) *Framework {
	return &Framework{
		key:              privKey,
		addressValidator: addressValidator,
		messageChan:      make(chan *IncomingMessage),
	}
}

// Framework is a framework for interacting with other peers using ethereum signatures and a kademlia style DHT,
// and only accepting messages from peers that are specified as valid.
type Framework struct {
	// Inherit methods we don't use
	network.GenericFramework

	// Used to check if an address is acceptable
	addressValidator func(string) bool

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

// Configure is used to set up our keystore, and block until we are ready to send/receive messages
func (f *Framework) Configure(l *network.Legion) {
	f.l = l
	id := &ID{
		EthAddress:     crypto.PubkeyToAddress(f.key.PublicKey).Bytes(),
		NetworkAddress: l.Me().String(),
	}
	f.self = id
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

	// Get the public key
	pubKey, err := crypto.SigToPub(hash, sm.Signature)
	if err != nil {
		return false
	}

	// Verify the signature
	if !crypto.VerifySignature(crypto.CompressPubkey(pubKey), hash, sm.Signature) {
		return false
	}

	// Check to see if the signed Ethereum Address matches sender
	m := &protobuf.DHTMessage{}
	err = m.Unmarshal(sm.DhtMessage)
	if err != nil {
		return false
	}

	// Make sure the sender matches the DHT message
	if !bytes.Equal(m.GetSender().EthAddress, crypto.PubkeyToAddress(*pubKey).Bytes()) {
		return false
	}

	// Validate that the sender network address matches what is signed
	if ctx.Sender.String() != m.GetSender().NetworkAddress {
		// Disconnect the peer
		ctx.Legion.DeletePeer(ctx.Sender)
		return false
	}

	// Finally check to see if the address is part of the pool
	return f.addressValidator(ID(*m.GetSender()).AddressHex())
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
func (f *Framework) SendMessage(recipient, messageType string, body proto.Message) error {
	return nil
}

// RecieveMessageChan returns a channel that recieves messages
func (f *Framework) RecieveMessageChan() chan *IncomingMessage {
	return f.messageChan
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
		// Preform lookup (send lookup request)
	} else if ctx.Message.Type == "dht.lookup_request" {
		// Respond with the closest nodes to the target
	} else if ctx.Message.Type == "dht.lookup_response" {
		// Verify response IDs are correct
	} else { // Send everything else to the recieve channel
		f.messageChan <- &IncomingMessage{Sender: dhtMessage.Sender, Body: dhtMessage}
	}
}

// PeerDisconnect is called when a peer is deleted
func (f *Framework) PeerDisconnect(ctx *network.PeerContext) {
	id, exists := f.idMap.Load(ctx.Peer.Remote())
	if exists {
		f.router.RemovePeer((id).(ID))
	}
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
