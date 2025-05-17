package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestRegisterInterfaces(t *testing.T) {
	interfaceRegistry := codectestutil.CodecOptions{}.NewInterfaceRegistry()
	RegisterInterfaces(interfaceRegistry)
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgFundCommunityPool{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCommunityPoolSpend{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCreateContinuousFund{}))
	require.NoError(t, interfaceRegistry.EnsureRegistered(&MsgCancelContinuousFund{}))
}

func TestNewMsgFundCommunityPool(t *testing.T) {
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	depositor := "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8"
	msg := NewMsgFundCommunityPool(amount, depositor)
	require.Equal(t, amount, msg.Amount)
	require.Equal(t, depositor, msg.Depositor)
}

func TestNewMsgCommunityPoolSpend(t *testing.T) {
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	authority := "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8"
	recipient := "cosmos1qypq2q2l8z4wz2z2l8z4wz2z2l8z4wz2z2l8z5"
	msg := NewMsgCommunityPoolSpend(amount, authority, recipient)
	require.Equal(t, amount, msg.Amount)
	require.Equal(t, authority, msg.Authority)
	require.Equal(t, recipient, msg.Recipient)
}

func TestValidateContinuousFund(t *testing.T) {
	testCases := []struct {
		name      string
		cf        ContinuousFund
		expErrMsg string
	}{
		{
			"valid continuous fund",
			ContinuousFund{
				Recipient:  "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8",
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
			"",
		},
		{
			"empty recipient",
			ContinuousFund{
				Recipient: "",
			},
			"recipient cannot be empty",
		},
		{
			"zero percentage",
			ContinuousFund{
				Recipient:  "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8",
				Percentage: math.LegacyZeroDec(),
			},
			"percentage cannot be zero or empty",
		},
		{
			"nil percentage",
			ContinuousFund{
				Recipient:  "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8",
				Percentage: math.LegacyDec{},
			},
			"percentage cannot be zero or empty",
		},
		{
			"negative percentage",
			ContinuousFund{
				Recipient:  "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8",
				Percentage: math.LegacyMustNewDecFromStr("-0.1"),
			},
			"percentage cannot be negative",
		},
		{
			"percentage exceeds 100%",
			ContinuousFund{
				Recipient:  "bc1pv3qgmvlr0gztdwn77a743cf4vu5s9nkyjsytf3vgezyan0lkaqjs0qz3z8",
				Percentage: math.LegacyMustNewDecFromStr("1.1"),
			},
			"percentage cannot be greater than one",
		},
	}

	for _, tc := range testCases {
		err := tc.cf.Validate()
		if tc.expErrMsg == "" {
			require.NoError(t, err)
		} else {
			require.EqualError(t, err, tc.expErrMsg)
		}
	}
}
