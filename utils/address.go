package utils

import (
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
	host := split[0]
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

func resolveHost(host string) string {
	return host
}
