package kava

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/proto"
)

type ClientStateI interface {
	proto.Message
}

// ClientState defines a solo machine client that tracks the current consensus
// state and if the client is frozen.
type ClientState struct {
	// latest sequence of the client state
	Sequence uint64 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	// frozen sequence of the solo machine
	IsFrozen       bool            `protobuf:"varint,2,opt,name=is_frozen,json=isFrozen,proto3" json:"is_frozen,omitempty" yaml:"is_frozen"`
	ConsensusState *ConsensusState `protobuf:"bytes,3,opt,name=consensus_state,json=consensusState,proto3" json:"consensus_state,omitempty" yaml:"consensus_state"`
	// when set to true, will allow governance to update a solo machine client.
	// The client will be unfrozen if it is frozen.
	AllowUpdateAfterProposal bool `protobuf:"varint,4,opt,name=allow_update_after_proposal,json=allowUpdateAfterProposal,proto3" json:"allow_update_after_proposal,omitempty" yaml:"allow_update_after_proposal"`
}

func (m *ClientState) Reset()         { *m = ClientState{} }
func (m *ClientState) String() string { return proto.CompactTextString(m) }
func (*ClientState) ProtoMessage()    {}

// ConsensusState defines a solo machine consensus state. The sequence of a
// consensus state is contained in the "height" key used in storing the
// consensus state.
type ConsensusState struct {
	// public key of the solo machine
	PublicKey *types.Any `protobuf:"bytes,1,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty" yaml:"public_key"`
	// diversifier allows the same public key to be re-used across different solo
	// machine clients (potentially on different chains) without being considered
	// misbehaviour.
	Diversifier string `protobuf:"bytes,2,opt,name=diversifier,proto3" json:"diversifier,omitempty"`
	Timestamp   uint64 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (m *ConsensusState) Reset()         { *m = ConsensusState{} }
func (m *ConsensusState) String() string { return proto.CompactTextString(m) }
func (*ConsensusState) ProtoMessage()    {}
