package address

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"
)

func TestTaprootCodec(t *testing.T) {
	mainnetParams := &chaincfg.MainNetParams
	testnetParams := &chaincfg.TestNet3Params

	tests := []struct {
		name        string
		network     *chaincfg.Params
		address     string
		expectError bool
	}{
		{
			name:        "valid mainnet taproot address",
			network:     mainnetParams,
			address:     "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0",
			expectError: false,
		},
		{
			name:        "valid mainnet address with other format",
			network:     mainnetParams,
			address:     "bc1qqx7a4rtcf7f49f8537xu9e2p0xt54k9jxlf6wm",
			expectError: true,
		},
		{
			name:        "valid testnet taproot address",
			network:     testnetParams,
			address:     "tb1prtlfl6ql8zgkrvak6c5xksa0tkmr524apnlxy3ut777qc9785vfsmuhg9g",
			expectError: false,
		},
		{
			name:        "invalid bech32 address",
			network:     mainnetParams,
			address:     "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4",
			expectError: true,
		},
		{
			name:        "invalid address format",
			network:     mainnetParams,
			address:     "not-an-address",
			expectError: true,
		},
		{
			name:        "empty address",
			network:     mainnetParams,
			address:     "",
			expectError: true,
		},
		{
			name:        "wrong network address",
			network:     mainnetParams,
			address:     "tb1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := NewTaprootCodec(tt.network)

			// Test StringToBytes
			bytes, err := codec.StringToBytes(tt.address)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, bytes)

			// Test BytesToString
			addrStr, err := codec.BytesToString(bytes)
			require.NoError(t, err)
			require.Equal(t, tt.address, addrStr)
		})
	}
}

func TestTaprootCodecRoundTrip(t *testing.T) {
	mainnetParams := &chaincfg.MainNetParams
	codec := NewTaprootCodec(mainnetParams)

	// Test round trip with a valid taproot address
	originalAddr := "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0"

	// Convert to bytes
	bytes, err := codec.StringToBytes(originalAddr)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	// Convert back to string
	addrStr, err := codec.BytesToString(bytes)
	require.NoError(t, err)
	require.Equal(t, originalAddr, addrStr)
}

func TestTaprootCodecEmptyBytes(t *testing.T) {
	mainnetParams := &chaincfg.MainNetParams
	codec := NewTaprootCodec(mainnetParams)

	// Test with empty bytes
	addrStr, err := codec.BytesToString([]byte{})
	require.NoError(t, err)
	require.Empty(t, addrStr)
}
