package utils

import (
	"strconv"
	"strings"
)

// NewLegionAddress returns a LegionAddress object
func NewLegionAddress(host string, port uint16) LegionAddress {
	host = resolveHost(host)
	return LegionAddress{host: host, port: port}
}

// FromString returns a LegionAddress from a string
func FromString(addrString string) LegionAddress {
	split := strings.Split(addrString, ":")
	host := split[0]
	port, _ := strconv.Atoi(split[1])
	return LegionAddress{host: host, port: uint16(port)}
}

//LegionAddress is a comparable type with a few convinience methods
type LegionAddress struct {
	host string
	port uint16
}

// String returns a formatted KCP address like "host:port"
func (k *LegionAddress) String() string {
	return k.host + ":" + strconv.Itoa(int(k.port))
}

func resolveHost(host string) string {
	return host
}
