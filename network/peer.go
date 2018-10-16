package network

import "github.com/gladiusio/legion/utils"

// Peer is an type that allows easy communication with
// a remote peer
type Peer struct {
	remote utils.KCPAddress
	// Need some KCP connection here
}

func (p *Peer) Write() {

}

func (p *Peer) Read() {

}

// Dial opens a stream to the peer
func (p *Peer) Dial() {

}

// Close closes the stream if it exists
func (p *Peer) Close() {

}
