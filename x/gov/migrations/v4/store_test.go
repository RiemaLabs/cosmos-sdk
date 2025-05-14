package v4_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v1gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	_, _, addr   = testdata.KeyTestPubAddr()
	govAcct      = authtypes.NewModuleAddress(types.ModuleName)
	TestProposal = getTestProposal()
)

type mockSubspace struct {
	dp v1.DepositParams
	vp v1.VotingParams
	tp v1.TallyParams
}

func newMockSubspace(p v1.Params) mockSubspace {
	return mockSubspace{
		dp: v1.DepositParams{
			MinDeposit:       p.MinDeposit,
			MaxDepositPeriod: p.MaxDepositPeriod,
		},
		vp: v1.VotingParams{
			VotingPeriod: p.VotingPeriod,
		},
		tp: v1.TallyParams{
			Quorum:        p.Quorum,
			Threshold:     p.Threshold,
			VetoThreshold: p.VetoThreshold,
		},
	}
}

func (ms mockSubspace) Get(ctx sdk.Context, key []byte, ptr any) {
	switch string(key) {
	case string(v1.ParamStoreKeyDepositParams):
		*ptr.(*v1.DepositParams) = ms.dp
	case string(v1.ParamStoreKeyVotingParams):
		*ptr.(*v1.VotingParams) = ms.vp
	case string(v1.ParamStoreKeyTallyParams):
		*ptr.(*v1.TallyParams) = ms.tp
	}
}

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	legacySubspace := newMockSubspace(v1.DefaultParams())

	propTime := time.Unix(1e9, 0)

	// Create 2 proposals
	prop1Content, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Test", "description"), authtypes.NewModuleAddress("gov").String())
	require.NoError(t, err)
	proposal1, err := v1.NewProposal([]sdk.Msg{prop1Content}, 1, propTime, propTime, "some metadata for the legacy content", "Test", "description", sdk.AccAddress([]byte{0xdd, 0x74, 0x7d, 0x28, 0xa0, 0x3f, 0xe2, 0x3c, 0xdd, 0xed, 0x64, 0x97, 0x6c, 0xc1, 0xf5, 0x32, 0x6b, 0x19, 0x8f, 0x20, 0xc2, 0x1c, 0x30, 0x9c, 0x3c, 0x5b, 0x8e, 0x62, 0x7c, 0xc5, 0x99, 0x69}), false)
	require.NoError(t, err)
	prop1Bz, err := cdc.Marshal(&proposal1)
	require.NoError(t, err)
	store.Set(v1gov.ProposalKey(proposal1.Id), prop1Bz)

	proposal2, err := v1.NewProposal(getTestProposal(), 2, propTime, propTime, "some metadata for the legacy content", "Test", "description", sdk.AccAddress([]byte{0x6c, 0x6a, 0x7f, 0x10, 0xe0, 0x67, 0xe, 0xd5, 0x6f, 0x1a, 0x4a, 0xf2, 0xc, 0x8a, 0xcb, 0xf6, 0xf4, 0x8a, 0x35, 0xb2, 0xe0, 0x5d, 0x96, 0x1d, 0xf6, 0x6b, 0x18, 0x2a, 0xd, 0xba, 0xf6, 0xad}), false)
	proposal2.Status = v1.StatusVotingPeriod
	require.NoError(t, err)
	prop2Bz, err := cdc.Marshal(&proposal2)
	require.NoError(t, err)
	store.Set(v1gov.ProposalKey(proposal2.Id), prop2Bz)

	// Run migrations.
	storeService := runtime.NewKVStoreService(govKey)
	err = v4.MigrateStore(ctx, storeService, legacySubspace, cdc)
	require.NoError(t, err)

	// Check params
	var params v1.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, legacySubspace.dp.MinDeposit, params.MinDeposit)
	require.Equal(t, legacySubspace.dp.MaxDepositPeriod, params.MaxDepositPeriod)
	require.Equal(t, legacySubspace.vp.VotingPeriod, params.VotingPeriod)
	require.Equal(t, legacySubspace.tp.Quorum, params.Quorum)
	require.Equal(t, legacySubspace.tp.Threshold, params.Threshold)
	require.Equal(t, legacySubspace.tp.VetoThreshold, params.VetoThreshold)
	require.Equal(t, math.LegacyZeroDec().String(), params.MinInitialDepositRatio)

	// Check proposals' status
	var migratedProp1 v1.Proposal
	bz = store.Get(v1gov.ProposalKey(proposal1.Id))
	require.NoError(t, cdc.Unmarshal(bz, &migratedProp1))
	require.Equal(t, v1.StatusDepositPeriod, migratedProp1.Status)

	var migratedProp2 v1.Proposal
	bz = store.Get(v1gov.ProposalKey(proposal2.Id))
	require.NoError(t, cdc.Unmarshal(bz, &migratedProp2))
	require.Equal(t, v1.StatusVotingPeriod, migratedProp2.Status)

	// Check if proposal 2 is in the new store but not proposal 1
	require.Nil(t, store.Get(v4.VotingPeriodProposalKey(proposal1.Id)))
	require.Equal(t, []byte{0x1}, store.Get(v4.VotingPeriodProposalKey(proposal2.Id)))
}

func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000)))),
		legacyProposalMsg,
	}
}
