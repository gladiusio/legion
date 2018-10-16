package utils

// NewKCPAddress returns a KCPAddress object
func NewKCPAddress(host string, port uint16) *KCPAddress {
	return &KCPAddress{host: host, port: port}
}

// KCPAddress is a wrapper that allows easy interaction and
// parsing of addresses, it also deals with address deduplication
// by returning the same string if addresses have the same
// actual destination
type KCPAddress struct {
	host string
	port uint16
}

// String returns a formatted KCP address
func (k KCPAddress) String() {

}
