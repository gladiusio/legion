package state

import "time"

// New creates a new state struct
func New(ip, port, currentMessage string) *State {
	return &State{ip: ip, port: port, currentMessage: currentMessage, creationTime: time.Now()}
}

// FromMessage creates a state struct based on the signed JSON string
func FromMessage(payload string) *State {
	return &State{}
}

// State represents current state of a node
type State struct {
	ip             string
	port           string
	currentMessage string // JSON String. This allows us to encode more information without updating the protocol
	creationTime   time.Time
	address        string
}

// IP - Gets the IP assosiated with the state
func (s State) IP() string {
	return s.ip
}

// Serialize returns the serialized (JSON) of the state
func (s State) Serialize() string {
	return ""
}
