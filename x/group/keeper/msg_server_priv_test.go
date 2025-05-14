package keeper

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/math"
)

func TestDoTallyAndUpdate(t *testing.T) {
	var (
		myAddr      = sdk.AccAddress([]byte{0xdd, 0x74, 0x7d, 0x28, 0xa0, 0x3f, 0xe2, 0x3c, 0xdd, 0xed, 0x64, 0x97, 0x6c, 0xc1, 0xf5, 0x32, 0x6b, 0x19, 0x8f, 0x20, 0xc2, 0x1c, 0x30, 0x9c, 0x3c, 0x5b, 0x8e, 0x62, 0x7c, 0xc5, 0x99, 0x69})
		myOtherAddr = sdk.AccAddress([]byte{0x6c, 0x6a, 0x7f, 0x10, 0xe0, 0x67, 0xe, 0xd5, 0x6f, 0x1a, 0x4a, 0xf2, 0xc, 0x8a, 0xcb, 0xf6, 0xf4, 0x8a, 0x35, 0xb2, 0xe0, 0x5d, 0x96, 0x1d, 0xf6, 0x6b, 0x18, 0x2a, 0xd, 0xba, 0xf6, 0xad})
	)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	myAccountKeeper := &mockAccountKeeper{
		AddressCodecFn: func() coreaddress.Codec {
			return address.NewTaprootCodec(&sdk.BitcoinNetParams)
		},
	}
	groupKeeper := NewKeeper(storeKey, encCfg.Codec, nil, myAccountKeeper, group.DefaultConfig())
	noEventsFn := func(proposalID uint64) sdk.Events { return sdk.Events{} }
	type memberVote struct {
		address string
		weight  string
		option  group.VoteOption
	}
	specs := map[string]struct {
		votes           []memberVote
		policy          group.DecisionPolicy
		expStatus       group.ProposalStatus
		expVotesCleared bool
		expEvents       func(proposalID uint64) sdk.Events
	}{
		"proposal accepted": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "2"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: true, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_ACCEPTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal rejected": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: true}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents:       noEventsFn,
		},
		"proposal in flight": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "1"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{Allow: false, Final: false}, nil
				},
			},
			expStatus:       group.PROPOSAL_STATUS_SUBMITTED,
			expVotesCleared: false,
			expEvents:       noEventsFn,
		},
		"policy errors": {
			votes: []memberVote{
				{address: myAddr.String(), option: group.VOTE_OPTION_YES, weight: "1"},
				{address: myOtherAddr.String(), option: group.VOTE_OPTION_NO, weight: "2"},
			},
			policy: mockDecisionPolicy{
				AllowFn: func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
					return group.DecisionPolicyResult{}, errors.New("my test error")
				},
			},
			expStatus:       group.PROPOSAL_STATUS_REJECTED,
			expVotesCleared: true,
			expEvents: func(proposalID uint64) sdk.Events {
				return sdk.Events{
					sdk.NewEvent("cosmos.group.v1.EventTallyError",
						sdk.Attribute{Key: "error_message", Value: `"my test error"`},
						sdk.Attribute{Key: "proposal_id", Value: fmt.Sprintf(`"%d"`, proposalID)},
					),
				}
			},
		},
	}
	var (
		groupID    uint64
		proposalId uint64
	)
	for name, spec := range specs {
		groupID++
		proposalId++
		t.Run(name, func(t *testing.T) {
			em := sdk.NewEventManager()
			ctx := testCtx.Ctx.WithEventManager(em)
			totalWeight, err := math.NewDecFromString("0")
			require.NoError(t, err)
			// given a group, policy and persisted votes
			for _, v := range spec.votes {
				err := groupKeeper.groupMemberTable.Create(ctx.KVStore(storeKey), &group.GroupMember{
					GroupId: groupID,
					Member:  &group.Member{Address: v.address, Weight: v.weight},
				})
				require.NoError(t, err)
				err = groupKeeper.voteTable.Create(ctx.KVStore(storeKey), &group.Vote{
					ProposalId: proposalId,
					Voter:      v.address,
					Option:     v.option,
				})
				require.NoError(t, err)
			}
			myGroupInfo := group.GroupInfo{
				TotalWeight: totalWeight.String(),
			}
			myPolicy := group.GroupPolicyInfo{GroupId: groupID}
			err = myPolicy.SetDecisionPolicy(spec.policy)
			require.NoError(t, err)

			myProposal := &group.Proposal{
				Id:              proposalId,
				Status:          group.PROPOSAL_STATUS_SUBMITTED,
				VotingPeriodEnd: ctx.BlockTime().Add(time.Hour),
			}

			// when
			gotErr := groupKeeper.doTallyAndUpdate(ctx, myProposal, myGroupInfo, myPolicy)
			// then
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expStatus, myProposal.Status)
			require.Equal(t, spec.expEvents(proposalId), em.Events())
			// and persistent state updated
			persistedVotes, err := groupKeeper.votesByProposal(ctx, groupID)
			require.NoError(t, err)
			if spec.expVotesCleared {
				assert.Empty(t, persistedVotes)
			} else {
				assert.Len(t, persistedVotes, len(spec.votes))
			}
		})
	}
}

var _ group.AccountKeeper = &mockAccountKeeper{}

// mockAccountKeeper is a mock implementation of the AccountKeeper interface for testing purposes.
type mockAccountKeeper struct {
	AddressCodecFn func() coreaddress.Codec
}

func (m mockAccountKeeper) AddressCodec() coreaddress.Codec {
	if m.AddressCodecFn == nil {
		panic("not expected to be called")
	}
	return m.AddressCodecFn()
}

func (m mockAccountKeeper) NewAccount(ctx context.Context, i sdk.AccountI) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) GetAccount(ctx context.Context, address sdk.AccAddress) sdk.AccountI {
	panic("not expected to be called")
}

func (m mockAccountKeeper) SetAccount(ctx context.Context, i sdk.AccountI) {
	panic("not expected to be called")
}

func (m mockAccountKeeper) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
	panic("not expected to be called")
}

// mockDecisionPolicy is a mock implementation of a decision policy for testing purposes.
type mockDecisionPolicy struct {
	fakeProtoType
	AllowFn func(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error)
}

func (m mockDecisionPolicy) Allow(tallyResult group.TallyResult, totalPower string) (group.DecisionPolicyResult, error) {
	if m.AllowFn == nil {
		panic("not expected to be called")
	}
	return m.AllowFn(tallyResult, totalPower)
}

func (m mockDecisionPolicy) GetVotingPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) GetMinExecutionPeriod() time.Duration {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) ValidateBasic() error {
	panic("not expected to be called")
}

func (m mockDecisionPolicy) Validate(g group.GroupInfo, config group.Config) error {
	panic("not expected to be called")
}

var (
	_ proto.Marshaler = (*fakeProtoType)(nil)
	_ proto.Message   = (*fakeProtoType)(nil)
)

// fakeProtoType is a struct used for mocking and testing purposes.
// Custom types can be converted into Any and back via internal CachedValue only.
type fakeProtoType struct{}

func (a fakeProtoType) Reset() {}

func (a fakeProtoType) String() string {
	return "testing"
}

func (a fakeProtoType) Marshal() ([]byte, error) {
	return nil, nil
}

func (a fakeProtoType) ProtoMessage() {}
