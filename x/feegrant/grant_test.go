package feegrant_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/module"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestGrant(t *testing.T) {
	addressCodec := codecaddress.NewTaprootCodec(&sdk.BitcoinNetParams)
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})

	addr, err := addressCodec.StringToBytes("bc1pzmgv90s54p22fjx68yftj8awm4psf5x6u85ces806pcnpt5rna0s4p9d7m")
	require.NoError(t, err)
	addr2, err := addressCodec.StringToBytes("bc1pej4p7zzqnjxfp7ut73azesfktn89fv6uz8ace3kp3au3vtaxk3yse5t4e4")
	require.NoError(t, err)
	atom := sdk.NewCoins(sdk.NewInt64Coin("atom", 555))
	now := ctx.BlockTime()
	oneYear := now.AddDate(1, 0, 0)

	zeroAtoms := sdk.NewCoins(sdk.NewInt64Coin("atom", 0))

	cases := map[string]struct {
		granter sdk.AccAddress
		grantee sdk.AccAddress
		limit   sdk.Coins
		expires time.Time
		valid   bool
	}{
		"good": {
			granter: addr2,
			grantee: addr,
			limit:   atom,
			expires: oneYear,
			valid:   true,
		},
		"no grantee": {
			granter: addr2,
			grantee: nil,
			limit:   atom,
			expires: oneYear,
			valid:   false,
		},
		"no granter": {
			granter: nil,
			grantee: addr,
			limit:   atom,
			expires: oneYear,
			valid:   false,
		},
		"self-grant": {
			granter: addr2,
			grantee: addr2,
			limit:   atom,
			expires: oneYear,
			valid:   false,
		},
		"zero allowance": {
			granter: addr2,
			grantee: addr,
			limit:   zeroAtoms,
			expires: oneYear,
			valid:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			grant, err := feegrant.NewGrant(tc.granter, tc.grantee, &feegrant.BasicAllowance{
				SpendLimit: tc.limit,
				Expiration: &tc.expires,
			})
			require.NoError(t, err)
			err = grant.ValidateBasic()

			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// if it is valid, let's try to serialize, deserialize, and make sure it matches
			bz, err := encCfg.Codec.Marshal(&grant)
			require.NoError(t, err)
			var loaded feegrant.Grant
			err = encCfg.Codec.Unmarshal(bz, &loaded)
			require.NoError(t, err)

			err = loaded.ValidateBasic()
			require.NoError(t, err)

			require.Equal(t, grant, loaded)
		})
	}
}
