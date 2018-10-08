package peer

import (
	"encoding/json"

	"github.com/gladiusio/gladius-controld/pkg/p2p/message"
	"github.com/gladiusio/gladius-controld/pkg/p2p/signature"
	"github.com/gladiusio/gladius-controld/pkg/p2p/state"
)

type delegate struct {
	peer *Peer
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

// NotifyMsg is called when a new message is recieved by this peer
func (d *delegate) NotifyMsg(b []byte) {
	var incommingUpdate *update
	if err := json.Unmarshal(b, &incommingUpdate); err != nil {
		panic(err)
	}
	switch incommingUpdate.Action {
	case "merge": // This is when a node is propigating a new message via gossip
		var sm *signature.SignedMessage

		err := json.Unmarshal([]byte(incommingUpdate.Data), &sm)
		if err != nil {
			panic(err)
		}
		go d.peer.GetState().UpdateState(sm)

	case "challenge_response": // This is from a node responding to a challenge question
		var sm *signature.SignedMessage

		smBytes, err := incommingUpdate.Data.MarshalJSON()
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(smBytes, &sm)
		if err != nil {
			panic(err)
		}

		mBytes, err := sm.Message.MarshalJSON()
		if err != nil {
			panic(err)
		}

		m := &message.Message{}

		err = json.Unmarshal(mBytes, m)
		if err != nil {
			panic(err)
		}

		c := &challenge{}
		cBytes, err := m.Content.MarshalJSON()
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(cBytes, c)
		if err != nil {
			panic(err)
		}

		// Queue up our incomming challenge
		channel, err := d.peer.getChallengeResponseChannel(c.ChallengeID)
		if err != nil {
			return
		}

		go func() { channel <- sm }()

	case "challenge_question": // This is a node recieving a question
		challengeBytes, err := incommingUpdate.Data.MarshalJSON()
		if err != nil {
			panic(err)
		}

		m := message.New(challengeBytes)
		sm, err := signature.CreateSignedMessage(m, d.peer.ga)
		if err != nil {
			panic(err)
		}
		smBytes, err := json.Marshal(sm)
		if err != nil {
			panic(err)
		}

		updateToReply := &update{
			From:   *d.peer.member.LocalNode(),
			Action: "challenge_response",
			Data:   smBytes,
		}

		updateBytes, err := json.Marshal(updateToReply)
		if err != nil {
			panic(err)
		}

		err = d.peer.member.SendReliable(&incommingUpdate.From, updateBytes)
		if err != nil {
			panic(err)
		}

	default:
		panic("unsupported update action")
	}

}

// GetBroadcasts returns the list of broadcast messages (not for us)
func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.peer.PeerQueue.GetBroadcasts(overhead, limit)
}

// Get the local state that we can pass on to another node for replication
func (d *delegate) LocalState(join bool) []byte {
	b, err := d.peer.GetState().GetJSON()

	if err != nil {
		panic(err)
	}

	return b
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	go func() {
		incomingState, err := state.ParseNetworkState(buf)
		if err != nil {
			panic(err)
		}
		// Get the signatures and rebuild the state
		sigList := incomingState.GetSignatureList()
		for _, sig := range sigList {
			d.peer.GetState().UpdateState(sig)
		}
	}()
}
