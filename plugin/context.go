package plugin

// MessageContext has context for a given message such as the legion object
// and methods to interact with the remote peer that sent the message
type MessageContext struct {
}

// PeerContext has context for a peer event such as the legion object and
// the peer change that fired the event
type PeerContext struct {
}

// NetworkContext is general context of the network, gives access to just
// the legion object and a few other helpers
type NetworkContext struct {
}
