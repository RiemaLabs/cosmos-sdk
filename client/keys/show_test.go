package keys

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/taproot"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func Test_multiSigKey_Properties(t *testing.T) {
	tmpKey1 := taproot.GenPrivKeyFromSecret([]byte("mySecret"))
	pk := multisig.NewLegacyAminoPubKey(
		1,
		[]cryptotypes.PubKey{tmpKey1.PubKey()},
	)
	k, err := keyring.NewMultiRecord("myMultisig", pk)
	require.NoError(t, err)
	require.Equal(t, "myMultisig", k.Name)
	require.Equal(t, keyring.TypeMulti, k.GetType())

	pub, err := k.GetPubKey()
	require.NoError(t, err)
	require.Equal(t, "0012DC9DD81459D21C363387A4B2EC8DE84A53003E2F3D3AF20A67EAA82CC977", pub.Address().String())

	addr, err := k.GetAddress()
	require.NoError(t, err)
	require.Equal(t, "cosmos1qqfde8wcz3vay8pkxwr6fvhv3h5y55cq8chn6whjpfn742pve9ms5ttpaq", sdk.MustBech32ifyAddressBytes("cosmos", addr))
}

func Test_showKeysCmd(t *testing.T) {
	cmd := ShowKeysCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "false", cmd.Flag(FlagAddress).DefValue)
	require.Equal(t, "false", cmd.Flag(FlagPublicKey).DefValue)
}

func Test_runShowCmd(t *testing.T) {
	cmd := ShowKeysCmd()
	cmd.Flags().AddFlagSet(Commands().PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	kbHome := t.TempDir()
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn, cdc)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithKeyringDir(kbHome).
		WithCodec(cdc)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{"invalid"})
	require.EqualError(t, cmd.ExecuteContext(ctx), "invalid is not a valid name or address: decoded address is of unknown format")

	cmd.SetArgs([]string{"invalid1", "invalid2"})
	require.EqualError(t, cmd.ExecuteContext(ctx), "invalid1 is not a valid name or address: decoded address is of unknown format")

	fakeKeyName1 := "runShowCmd_Key1"
	fakeKeyName2 := "runShowCmd_Key2"

	t.Cleanup(func() {
		cleanupKeys(t, kb, "runShowCmd_Key1")
		cleanupKeys(t, kb, "runShowCmd_Key2")
	})

	path := hd.NewFundraiserParams(1, sdk.CoinType, 0).String()
	_, err = kb.NewAccount(fakeKeyName1, testdata.TestMnemonic, "", path, hd.Taproot)
	require.NoError(t, err)

	path2 := hd.NewFundraiserParams(1, sdk.CoinType, 1).String()
	_, err = kb.NewAccount(fakeKeyName2, testdata.TestMnemonic, "", path2, hd.Taproot)
	require.NoError(t, err)

	// Now try single key
	cmd.SetArgs([]string{
		fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=", FlagBechPrefix),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.EqualError(t, cmd.ExecuteContext(ctx), "invalid Bech32 prefix encoding provided: ")

	cmd.SetArgs([]string{
		fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	// try fetch by name
	require.NoError(t, cmd.ExecuteContext(ctx))

	// try fetch by addr
	k, err := kb.Key(fakeKeyName1)
	require.NoError(t, err)
	addr, err := k.GetAddress()
	require.NoError(t, err)
	cmd.SetArgs([]string{
		addr.String(),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	require.NoError(t, cmd.ExecuteContext(ctx))

	// Now try multisig key - set bech to acc
	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
		fmt.Sprintf("--%s=0", flagMultiSigThreshold),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.EqualError(t, cmd.ExecuteContext(ctx), "threshold must be a positive integer")

	// Now try multisig key duplicate
	_, mockOut := testutil.ApplyMockIO(cmd)
	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))
	require.Contains(t, mockOut.String(), fmt.Sprintf("WARNING: duplicate keys found: %s", fakeKeyName1))

	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=%s", FlagBechPrefix, sdk.PrefixAccount),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.NoError(t, cmd.ExecuteContext(ctx))

	// Now try multisig key - set bech to acc + threshold=2
	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=acc", FlagBechPrefix),
		fmt.Sprintf("--%s=true", FlagDevice),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.EqualError(t, cmd.ExecuteContext(ctx), "the device flag (-d) can only be used for accounts stored in devices")

	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=val", FlagBechPrefix),
		fmt.Sprintf("--%s=true", FlagDevice),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.EqualError(t, cmd.ExecuteContext(ctx), "the device flag (-d) can only be used for accounts")

	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=val", FlagBechPrefix),
		fmt.Sprintf("--%s=true", FlagDevice),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=true", FlagPublicKey),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})
	require.EqualError(t, cmd.ExecuteContext(ctx), "the device flag (-d) can only be used for addresses not pubkeys")

	cmd.SetArgs([]string{
		fakeKeyName1,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=true", FlagAddress),
		fmt.Sprintf("--%s=true", flagQRCode),
		// we have to reset following flags as they were set to true above, and won't be auto reset to false if we skip to specify these flags.
		// Note: this maybe a bug about spf13/cobra as cmd.flags's value won't be reset by changing cmd.args with cmd.SetArgs.
		fmt.Sprintf("--%s=false", FlagDevice),
		fmt.Sprintf("--%s=false", FlagPublicKey),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	// try fetch by name
	require.NoError(t, cmd.ExecuteContext(ctx))

	cmd.SetArgs([]string{
		fakeKeyName1, fakeKeyName2,
		fmt.Sprintf("--%s=%s", flags.FlagKeyringDir, kbHome),
		fmt.Sprintf("--%s=true", FlagAddress),
		fmt.Sprintf("--%s=true", flagQRCode),
		fmt.Sprintf("--%s=2", flagMultiSigThreshold),
		fmt.Sprintf("--%s=false", FlagDevice),
		fmt.Sprintf("--%s=false", FlagPublicKey),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	})

	// try fetch by name
	require.NoError(t, cmd.ExecuteContext(ctx))
}

func Test_validateMultisigThreshold(t *testing.T) {
	type args struct {
		k     int
		nKeys int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"zeros", args{0, 0}, true},
		{"1-0", args{1, 0}, true},
		{"1-1", args{1, 1}, false},
		{"1-2", args{1, 1}, false},
		{"1-2", args{2, 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMultisigThreshold(tt.args.k, tt.args.nKeys); (err != nil) != tt.wantErr {
				t.Errorf("validateMultisigThreshold() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getBechKeyOut(t *testing.T) {
	type args struct {
		bechPrefix string
	}
	tests := []struct {
		name    string
		args    args
		want    bechKeyOutFn
		wantErr bool
	}{
		{"empty", args{""}, nil, true},
		{"wrong", args{"???"}, nil, true},
		{"acc", args{sdk.PrefixAccount}, MkAccKeyOutput, false},
		{"val", args{sdk.PrefixValidator}, MkValKeyOutput, false},
		{"cons", args{sdk.PrefixConsensus}, MkConsKeyOutput, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getBechKeyOut(tt.args.bechPrefix)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}
