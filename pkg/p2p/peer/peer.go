package peer

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/gladiusio/gladius-controld/pkg/blockchain"
	"github.com/gladiusio/gladius-controld/pkg/p2p/signature"
	"github.com/gladiusio/gladius-controld/pkg/p2p/state"
	"github.com/hashicorp/memberlist"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// New returns a new peer type
func New(ga *blockchain.GladiusAccountManager) *Peer {
	d := &delegate{}
	md := &mergeDelegate{}
	hostname, _ := os.Hostname()

	c := memberlist.DefaultWANConfig()
	c.PushPullInterval = 60 * time.Second
	c.GossipInterval = 200 * time.Millisecond
	c.ProbeTimeout = 4 * time.Second
	c.ProbeInterval = 7 * time.Second
	c.GossipNodes = 5
	c.Delegate = d
	// FIXME: Renable this feature, problem now is that the challenges that nodes
	// respond with seem to be wrong
	// c.Merge = md
	c.Name = hostname + "-" + uuid.NewV4().String()
	c.AdvertisePort = viper.GetInt("P2P.AdvertisePort")
	c.BindPort = viper.GetInt("P2P.BindPort")

	m, err := memberlist.Create(c)
	if err != nil {
		panic(err)
	}

	queue := &memberlist.TransmitLimitedQueue{
		RetransmitMult: 4,
	}

	peer := &Peer{
		peerState:           &state.State{},
		running:             false,
		peerDelegate:        d,
		member:              m,
		PeerQueue:           queue,
		challengeReceiveMap: make(map[string]chan *signature.SignedMessage),
		ga:                  ga,
	}

	queue.NumNodes = func() int { return peer.member.NumMembers() }
	d.peer = peer
	md.peer = peer
	return peer
}

// Peer is a type that represents a peer in the Gladius p2p network.
type Peer struct {
	ga                  *blockchain.GladiusAccountManager
	peerDelegate        *delegate
	PeerQueue           *memberlist.TransmitLimitedQueue
	peerState           *state.State
	member              *memberlist.Memberlist
	running             bool
	challengeReceiveMap map[string]chan *signature.SignedMessage // Map of challenge set ids to a receive channel of the responses from the questioned peers.
	mux                 sync.Mutex
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type update struct {
	From   memberlist.Node
	Action string          // Can be "merge", "challenge_question", or "challenge_response"
	Data   json.RawMessage // Usually a signed message, but can also be a challenge question
}

// Used to send to a node through an "update"
type challenge struct {
	ChallengeID string
	Question    string
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

// Join will request to join the network from a specific node
func (p *Peer) Join(ipList []string) error {
	_, err := p.member.Join(ipList)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) SetState(s *state.State) {
	p.mux.Lock()
	p.peerState = s
	p.mux.Unlock()
}

// StopAndLeave will infomr the network of it leaving and shutdown
func (p *Peer) StopAndLeave() error {
	err := p.member.Leave(1 * time.Second)
	if err != nil {
		return err
	}

	err = p.member.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) registerOutgoingChallenge(challengeID string) {
	p.mux.Lock()
	p.challengeReceiveMap[challengeID] = make(chan *signature.SignedMessage)
	p.mux.Unlock()
}

func (p *Peer) getChallengeResponseChannel(challengeID string) (chan *signature.SignedMessage, error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	if challengeChan, ok := p.challengeReceiveMap[challengeID]; ok {
		return challengeChan, nil
	}
	return nil, errors.New("Could not find channel")
}

// UpdateAndPushState updates the local state and pushes it to several other peers
func (p *Peer) UpdateAndPushState(sm *signature.SignedMessage) error {
	err := p.GetState().UpdateState(sm)
	if err != nil {
		return err
	}

	signedBytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	b, err := json.Marshal(&update{
		Action: "merge",
		Data:   signedBytes,
		From:   *p.member.LocalNode(),
	})

	if err != nil {
		return err
	}

	p.PeerQueue.QueueBroadcast(&broadcast{
		msg:    b,
		notify: nil,
	})

	return nil
}

// GetState returns the current local state
func (p *Peer) GetState() *state.State {
	return p.peerState
}

// CompareContent compares the content provided with the content in the state
// and returns a list of the missing files names in the format of:
// website/<"asset" or "route">/filename
func (p *Peer) CompareContent(contentList []string) []interface{} {
	// Convert to an interface array
	cl := make([]interface{}, len(contentList))
	for i, v := range contentList {
		cl[i] = v
	}
	contentWeHaveSet := mapset.NewSetFromSlice(cl)

	contentField := p.GetState().GetPoolField("RequiredContent")
	if contentField == nil {
		return make([]interface{}, 0)
	}
	contentFromPool := contentField.(state.SignedList).Data

	// Convert to an interface array
	s := make([]interface{}, len(contentFromPool))
	for i, v := range contentFromPool {
		s[i] = v
	}

	// Create a set
	contentWeNeed := mapset.NewSetFromSlice(s)

	// Return the difference of the two
	return contentWeNeed.Difference(contentWeHaveSet).ToSlice()
}

// GetContentLinks returns a map mapping a file name to all the places it can
// be found on the network
func (p *Peer) GetContentLinks(contentList []string) map[string][]string {
	allContent := p.GetState().GetNodeFieldsMap("DiskContent")
	toReturn := make(map[string][]string)
	for nodeAddress, diskContent := range allContent {
		ourContent := diskContent.(state.SignedList).Data
		// Convert to an interface array
		s := make([]interface{}, len(ourContent))
		for i, v := range ourContent {
			s[i] = v
		}
		ourContentSet := mapset.NewSetFromSlice(s)
		// Check to see if the current node we're iterating over has any of the
		// content we want
		for _, contentWanted := range contentList {
			if ourContentSet.Contains(contentWanted) {
				if toReturn[contentWanted] == nil {
					toReturn[contentWanted] = make([]string, 0)
				}
				// Add the URL to the map
				link := p.createContentLink(nodeAddress, contentWanted)
				if link != "" {
					toReturn[contentWanted] = append(toReturn[contentWanted], link)
				}
			}
		}
	}
	return toReturn
}

// Builds a URL to a node
func (p *Peer) createContentLink(nodeAddress, contentFileName string) string {
	nodeIP := p.GetState().GetNodeField(nodeAddress, "IPAddress").(state.SignedField).Data
	nodePort := p.GetState().GetNodeField(nodeAddress, "ContentPort").(state.SignedField).Data

	contentData := strings.Split(contentFileName, "/")
	u := url.URL{}
	if nodeIP == nil || nodePort == nil {
		return ""
	}
	u.Host = nodeIP.(string) + ":" + nodePort.(string)
	u.Path = "/content"
	u.Scheme = "http"

	if len(contentData) == 2 {
		q := u.Query()
		q.Add("website", contentData[0]) // website name
		q.Add("asset", contentData[1])   // "asset" to name of file
		u.RawQuery = q.Encode()
		return u.String()
	}
	return ""
}
