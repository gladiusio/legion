package utils

import (
	"net"
	"strconv"
	"strings"
)

// NewLegionAddress returns a LegionAddress object
func NewLegionAddress(host string, port uint16) LegionAddress {
	host = resolveHost(host)
	return LegionAddress{Host: host, Port: port}
}

// LegionAddressFromString returns a LegionAddress from a string
func LegionAddressFromString(addrString string) LegionAddress {
	split := strings.Split(addrString, ":")
	if len(split) != 2 {
		return LegionAddress{}
	}
	host := split[0]
	host = resolveHost(host)
	port, _ := strconv.Atoi(split[1])
	return LegionAddress{Host: host, Port: uint16(port)}
}

//LegionAddress is a comparable type with a few convinience methods
type LegionAddress struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// String returns a formatted KCP address like "host:port"
func (k LegionAddress) String() string {
	return k.Host + ":" + strconv.Itoa(int(k.Port))
}

// IsValid returns true if the address is valid, false if not
func (k LegionAddress) IsValid() bool {
	if k.Host != "" && k.Port != 0 {
		return true
	}
	return false
}

func resolveHost(host string) string {
	addr, err := net.LookupHost(host)
	if len(addr) == 0 || err != nil {
		return host
	}
	return addr[0]
}
