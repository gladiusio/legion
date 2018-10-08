package signature

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	response2 "github.com/gladiusio/gladius-controld/pkg/routing/response"
	"github.com/gladiusio/gladius-controld/pkg/utils"
	"github.com/spf13/viper"

	"github.com/buger/jsonparser"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gladiusio/gladius-controld/pkg/blockchain"
	"github.com/gladiusio/gladius-controld/pkg/p2p/message"
	"github.com/gladiusio/gladius-utils/config"
	"github.com/tdewolff/minify"
	mjson "github.com/tdewolff/minify/json"
)

// SignedMessage is a type representing a signed message
type SignedMessage struct {
	Message   *json.RawMessage `json:"message"`
	Hash      []byte           `json:"hash"`
	Signature []byte           `json:"signature"`
	Address   string           `json:"address"`
	verified  bool             // TODO: Make this useful
}

// ParseSignedMessage returns a signed message to be passed into the VerifySignedMessage method
func ParseSignedMessage(message, hash, signature, address string) (*SignedMessage, error) {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), mjson.Minify)
	messageMin, err := m.String("text/json", message)
	if err != nil {
		panic(err)
	}

	h := json.RawMessage(messageMin)
	dHash, err := b64.StdEncoding.DecodeString(hash)
	if err != nil {
		return nil, errors.New("error decoding hash")
	}
	dSignature, err := b64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, errors.New("error decoding signature")
	}

	return &SignedMessage{Message: &h, Hash: dHash, Signature: dSignature, Address: address, verified: false}, nil
}

// GetTimestamp gets the verified timestamp from the message
func (sm SignedMessage) GetTimestamp() int64 {
	jsonBytes, _ := sm.Message.MarshalJSON()
	timestamp, _ := jsonparser.GetInt(jsonBytes, "timestamp")
	return timestamp
}

func (sm SignedMessage) GetAgeInSeconds() int64 {
	jsonBytes, _ := sm.Message.MarshalJSON()
	timestamp, _ := jsonparser.GetInt(jsonBytes, "timestamp")

	now := time.Now().Unix()

	return now - timestamp
}

// IsVerified checks the internal status of the message and returns true if the
// message is verified
func (sm SignedMessage) IsVerified() bool {
	// Check if hash matches the message
	m, _ := sm.Message.MarshalJSON()
	hashMatches := bytes.Equal(sm.Hash, crypto.Keccak256(m))

	pub, err := crypto.SigToPub(sm.Hash, sm.Signature)
	if err != nil {
		return false
	}

	// Check if the signature is valid
	signatureValid := crypto.VerifySignature(crypto.CompressPubkey(pub), sm.Hash, sm.Signature[:64])

	// Check if the address matches
	addressMatches := crypto.PubkeyToAddress(*pub).String() == sm.Address

	if addressMatches && hashMatches && signatureValid {
		return true
	}

	return false

}

func (sm SignedMessage) IsPoolManagerAndVerified() bool {
	return sm.IsVerified() && sm.Address == config.GetString("blockchain.poolManagerAddress")
}

func (sm SignedMessage) IsInPoolAndVerified() bool {
	// Check if address is part of pool
	// config override
	if viper.GetBool("P2P.VerifyOverride") {
		return sm.IsVerified() && true
	}

	if !sm.IsVerified() {
		return false
	}
	nodeAddress := sm.Address

	poolUrl := viper.GetString("Blockchain.PoolUrl")

	response, _ := utils.SendRequest(http.MethodGet, poolUrl+"applications/pool/contains/"+nodeAddress, nil)
	var defaultResponse response2.DefaultResponse
	json.Unmarshal([]byte(response), &defaultResponse)

	byteResponse, _ := json.Marshal(defaultResponse.Response)
	var poolContainsWallet struct{ ContainsWallet bool }
	json.Unmarshal(byteResponse, &poolContainsWallet)

	return poolContainsWallet.ContainsWallet
}

func CreateSignedMessage(message *message.Message, ga *blockchain.GladiusAccountManager) (*SignedMessage, error) {

	// Create a serialized JSON string
	messageBytes := message.Serialize()

	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), mjson.Minify)
	messageBytes, err := m.Bytes("text/json", messageBytes)
	if err != nil {
		panic(err)
	}

	hash := crypto.Keccak256(messageBytes)
	account, err := ga.GetAccount()

	if err != nil {
		return nil, err
	}

	signature, err := ga.Keystore().SignHash(*account, hash)
	if err != nil {
		return &SignedMessage{}, errors.New("Error signing message, wallet likely not unlocked")
	}

	address, err := ga.GetAccountAddress()
	if err != nil {
		return nil, err
	}

	addressString := address.String()

	h := json.RawMessage(messageBytes)

	// Create the signed message
	signed := &SignedMessage{
		Message:   &h,
		Hash:      hash,
		Signature: signature,
		Address:   addressString,
	}

	return signed, nil
}

// CreateSignedMessageString creates a signed state from the message where
func CreateSignedMessageString(message *message.Message, ga *blockchain.GladiusAccountManager) (string, error) {
	signed, err := CreateSignedMessage(message, ga)
	if err != nil {
		return "", err
	}
	// Encode the struct as a json
	bytes, err := json.Marshal(signed)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
