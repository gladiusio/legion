package ethpool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/bits"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gladiusio/legion/frameworks/ethpool/protobuf"
)

// ID wraps helper methods around the protobuf type
type ID protobuf.ID

// CreateID is a factory function creating ID.
func CreateID(address common.Address, networkAddr string) ID {
	return ID{EthAddress: address.Bytes(), NetworkAddress: networkAddr}
}

// String returns the ethereum address and network address
func (id ID) String() string {
	return fmt.Sprintf("ID{EthAddr: %v, NetAddr: %v}", id.EthAddress, id.NetworkAddress)
}

// Equals returns true if both have the same ethereum address and network address
func (id ID) Equals(other ID) bool {
	return bytes.Equal(id.EthAddress, other.EthAddress) && id.NetworkAddress == other.NetworkAddress
}

// Less determines if this peer ID's ethereum address is less than other ID's ethereum address.
func (id ID) Less(other interface{}) bool {
	if other, is := other.(ID); is {
		return bytes.Compare(id.EthAddress, other.EthAddress) == -1
	}
	return false
}

// AddressHex generates a hex-encoded string of the address.
func (id ID) AddressHex() string {
	return hex.EncodeToString(id.EthAddress)
}

// EthereumAddress returns the Ethereum address representation of the ID
func (id ID) EthereumAddress() common.Address {
	return common.HexToAddress(id.AddressHex())
}

// Xor performs XOR (^) over another peer ID's public key.
func (id ID) Xor(other ID) ID {
	result := make([]byte, len(id.EthAddress))

	for i := 0; i < len(id.EthAddress) && i < len(other.EthAddress); i++ {
		result[i] = id.EthAddress[i] ^ other.EthAddress[i]
	}
	return ID{NetworkAddress: id.NetworkAddress, EthAddress: result}
}

// PrefixLen returns the number of prefixed zeros in a peer ID.
func (id ID) PrefixLen() int {
	for i, b := range id.EthAddress {
		if b != 0 {
			return i*8 + bits.LeadingZeros8(uint8(b))
		}
	}
	return len(id.EthAddress)*8 - 1
}
