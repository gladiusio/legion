package events

// MessageEvent represents some sort of message event,
// like a new message
type MessageEvent int

// The actual events
const (
	NewMessageEvent MessageEvent = iota
)

// PeerEvent represents a peer event, like a new
// peer addition
type PeerEvent int

// The actual events
const (
	PeerAddEvent PeerEvent = iota
	PeerDeleteEvent
	PeerPromotionEvent
)

// NetworkEvent represents some sort of network event,
// like startup/shutdown.
type NetworkEvent int

// The actual events
const (
	StartupEvent NetworkEvent = iota
	CloseEvent
)
