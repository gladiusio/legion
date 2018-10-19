package utils

// NewKCPAddress returns a KCPAddress object
func NewKCPAddress(host string, port uint16) KCPAddress {
	host = resolveHost(host)
	return KCPAddress{host: host, port: port}
}

//KCPAddress is a comparable type with a few convinience methods
type KCPAddress struct {
	host string
	port uint16
}

// String returns a formatted KCP address like host:port
func (k KCPAddress) String() string {
	return ""
}

func resolveHost(host string) string {
	return ""
}
