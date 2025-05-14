package v1_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var (
	coinsPos   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsMulti = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
	addrs      = []sdk.AccAddress{
		sdk.AccAddress([]byte{0xdd, 0x74, 0x7d, 0x28, 0xa0, 0x3f, 0xe2, 0x3c, 0xdd, 0xed, 0x64, 0x97, 0x6c, 0xc1, 0xf5, 0x32, 0x6b, 0x19, 0x8f, 0x20, 0xc2, 0x1c, 0x30, 0x9c, 0x3c, 0x5b, 0x8e, 0x62, 0x7c, 0xc5, 0x99, 0x69}),
		sdk.AccAddress([]byte{0x6c, 0x6a, 0x7f, 0x10, 0xe0, 0x67, 0xe, 0xd5, 0x6f, 0x1a, 0x4a, 0xf2, 0xc, 0x8a, 0xcb, 0xf6, 0xf4, 0x8a, 0x35, 0xb2, 0xe0, 0x5d, 0x96, 0x1d, 0xf6, 0x6b, 0x18, 0x2a, 0xd, 0xba, 0xf6, 0xad}),
	}
)

func init() {
	coinsMulti.Sort()
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress([]byte{0xdd, 0x74, 0x7d, 0x28, 0xa0, 0x3f, 0xe2, 0x3c, 0xdd, 0xed, 0x64, 0x97, 0x6c, 0xc1, 0xf5, 0x32, 0x6b, 0x19, 0x8f, 0x20, 0xc2, 0x1c, 0x30, 0x9c, 0x3c, 0x5b, 0x8e, 0x62, 0x7c, 0xc5, 0x99, 0x69})
	msg := v1.NewMsgDeposit(addr, 0, coinsPos)
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	res, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)
	expected := `{"type":"cosmos-sdk/v1/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"bc1pm468629q8l3reh0dvjtkes04xf43nreqcgwrp8putw8xylx9n95s0alkjs","proposal_id":"0"}}`
	require.Equal(t, expected, string(res))
}

// this tests that Amino JSON MsgSubmitProposal.GetSignBytes() still works with Content as Any using the ModuleCdc
func TestMsgSubmitProposal_GetSignBytes(t *testing.T) {
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	testcases := []struct {
		name      string
		proposal  []sdk.Msg
		title     string
		summary   string
		expedited bool
		expSignBz string
	}{
		{
			"MsgVote",
			[]sdk.Msg{v1.NewMsgVote(addrs[0], 1, v1.OptionYes, "")},
			"gov/MsgVote",
			"Proposal for a governance vote msg",
			false,
			`{"type":"cosmos-sdk/v1/MsgSubmitProposal","value":{"initial_deposit":[],"messages":[{"type":"cosmos-sdk/v1/MsgVote","value":{"option":1,"proposal_id":"1","voter":"bc1pm468629q8l3reh0dvjtkes04xf43nreqcgwrp8putw8xylx9n95s0alkjs"}}],"summary":"Proposal for a governance vote msg","title":"gov/MsgVote"}}`,
		},
		{
			"MsgSend",
			[]sdk.Msg{banktypes.NewMsgSend(addrs[0], addrs[0], sdk.NewCoins())},
			"bank/MsgSend",
			"Proposal for a bank msg send",
			false,
			fmt.Sprintf(`{"type":"cosmos-sdk/v1/MsgSubmitProposal","value":{"initial_deposit":[],"messages":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[],"from_address":"%s","to_address":"%s"}}],"summary":"Proposal for a bank msg send","title":"bank/MsgSend"}}`, addrs[0], addrs[0]),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := v1.NewMsgSubmitProposal(tc.proposal, sdk.NewCoins(), sdk.AccAddress{}.String(), "", tc.title, tc.summary, tc.expedited)
			require.NoError(t, err)
			bz, err := pc.MarshalAminoJSON(msg)
			require.NoError(t, err)
			require.Equal(t, tc.expSignBz, string(bz))
		})
	}
}
