package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinsPos   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsMulti = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
)

func init() {
	coinsMulti.Sort()
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress([]byte{0xdd, 0x74, 0x7d, 0x28, 0xa0, 0x3f, 0xe2, 0x3c, 0xdd, 0xed, 0x64, 0x97, 0x6c, 0xc1, 0xf5, 0x32, 0x6b, 0x19, 0x8f, 0x20, 0xc2, 0x1c, 0x30, 0x9c, 0x3c, 0x5b, 0x8e, 0x62, 0x7c, 0xc5, 0x99, 0x69})
	msg := NewMsgDeposit(addr, 0, coinsPos)
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	res, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)

	expected := `{"type":"cosmos-sdk/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"bc1pm468629q8l3reh0dvjtkes04xf43nreqcgwrp8putw8xylx9n95s0alkjs","proposal_id":"0"}}`
	require.Equal(t, expected, string(res))
}

// this tests that Amino JSON MsgSubmitProposal.GetSignBytes() still works with Content as Any using the ModuleCdc
func TestMsgSubmitProposal_GetSignBytes(t *testing.T) {
	msg, err := NewMsgSubmitProposal(NewTextProposal("test", "abcd"), sdk.NewCoins(), sdk.AccAddress{})
	require.NoError(t, err)
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	bz, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)
	require.Equal(t,
		`{"type":"cosmos-sdk/MsgSubmitProposal","value":{"content":{"type":"cosmos-sdk/TextProposal","value":{"description":"abcd","title":"test"}},"initial_deposit":[]}}`,
		string(bz))
}
