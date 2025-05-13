package simapp

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	codecAddress "github.com/cosmos/cosmos-sdk/codec/address"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  logger.With("instance", "first"),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// BlockedAddresses returns a map of addresses in app v1 and a map of modules name in app di.
	for acc := range BlockedAddresses() {
		var addr sdk.AccAddress
		if modAddr, err := sdk.AccAddressFromBech32(acc); err == nil {
			addr = modAddr
		} else {
			addr = app.AccountKeeper.GetModuleAddress(acc)
		}

		require.True(
			t,
			app.BankKeeper.BlockedAddr(addr),
			fmt.Sprintf("ensure that blocked addresses are properly set in bank keeper: %s should be blocked", acc),
		)
	}

	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewSimApp(logger.With("instance", "second"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  log.NewTestLogger(t),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false)
	vm, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	for v, i := range app.ModuleManager.Modules {
		if i, ok := i.(module.HasConsensusVersion); ok {
			require.Equal(t, vm[v], i.ConsensusVersion())
		}
	}

	require.NotNil(t, app.UpgradeKeeper.GetVersionSetter())
}

// TestMergedRegistry tests that fetching the gogo/protov2 merged registry
// doesn't fail after loading all file descriptors.
func TestMergedRegistry(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	require.Greater(t, r.NumFiles(), 0)
}

func TestProtoAnnotations(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	err = msgservice.ValidateProtoAnnotations(r)
	require.NoError(t, err)
}

var _ address.Codec = (*customAddressCodec)(nil)

type customAddressCodec struct{}

func (c customAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (c customAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func TestAddressCodecFactory(t *testing.T) {
	var addrCodec address.Codec
	var valAddressCodec runtime.ValidatorAddressCodec
	var consAddressCodec runtime.ConsensusAddressCodec

	err := depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(log.NewNopLogger()),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok := addrCodec.(codecAddress.TaprootCodec)
	require.True(t, ok)
	_, ok = addrCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.False(t, ok)

	// Set the address codec to the custom one
	err = depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
				func() address.Codec { return customAddressCodec{} },
				func() runtime.ValidatorAddressCodec { return customAddressCodec{} },
				func() runtime.ConsensusAddressCodec { return customAddressCodec{} },
			),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok = addrCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.True(t, ok)
}
