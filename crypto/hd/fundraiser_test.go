package hd_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/taproot"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type addrData struct {
	Mnemonic string
	Master   string
	Seed     string
	Priv     string
	Pub      string
	Addr     string
}

func TestFullFundraiserPath(t *testing.T) {
	require.Equal(t, "m/44'/118'/0'/0/0", hd.NewFundraiserParams(0, 118, 0).String())
}

func initFundraiserTestVectors(t *testing.T) []addrData {
	t.Helper()

	// NOTE: atom fundraiser address
	// var hdPath string = "m/44'/118'/0'/0/0"
	var hdToAddrTable []addrData

	b, err := os.ReadFile("testdata/test.json")
	if err != nil {
		t.Fatalf("could not read fundraiser test vector file (testdata/test.json): %s", err)
	}

	err = json.Unmarshal(b, &hdToAddrTable)
	if err != nil {
		t.Fatalf("could not decode test vectors (testdata/test.json): %s", err)
	}
	return hdToAddrTable
}

func TestFundraiserCompatibility(t *testing.T) {
	hdToAddrTable := initFundraiserTestVectors(t)

	for i, d := range hdToAddrTable {
		privB, _ := hex.DecodeString(d.Priv)
		pubB, _ := hex.DecodeString(d.Pub)
		addrB, _ := hex.DecodeString(d.Addr)
		seedB, _ := hex.DecodeString(d.Seed)
		masterB, _ := hex.DecodeString(d.Master)

		seed := bip39.NewSeed(d.Mnemonic, "")

		t.Log("================================")
		t.Logf("ROUND: %d MNEMONIC: %s", i, d.Mnemonic)

		master, ch := hd.ComputeMastersFromSeed(seed)
		priv, err := hd.DerivePrivateKeyForPath(master, ch, "m/44'/118'/0'/0/0")
		require.NoError(t, err)

		privKey := &taproot.PrivKey{Key: priv}
		pub := privKey.PubKey()

		t.Log("\tNODEJS GOLANG\n")
		t.Logf("SEED \t%X %X\n", seedB, seed)
		t.Logf("MSTR \t%X %X\n", masterB, master)
		t.Logf("PRIV \t%X %X\n", privB, priv)
		t.Logf("PUB  \t%X %X\n", pubB, pub)

		require.Equal(t, seedB, seed)
		require.Equal(t, master[:], masterB, fmt.Sprintf("Expected masters to match for %d", i))
		require.Equal(t, priv, privB, "Expected priv keys to match")
		pubBFixed := make([]byte, taproot.PubKeySize)
		copy(pubBFixed, pubB)
		require.Equal(t, pub, &taproot.PubKey{Key: pubBFixed}, fmt.Sprintf("Expected pub keys to match for %d", i))

		addr := pub.Address()
		t.Logf("ADDR  \t%X %X\n", addrB, addr)
		require.Equal(t, addr, crypto.Address(addrB), fmt.Sprintf("Expected addresses to match %d", i))
	}
}

func TestBitcoinWalletCompatibility(t *testing.T) {

	testMnemonics := []string{
		"episode", "tip", "cereal", "rice", "subway", "pelican",
		"family", "advice", "profit", "pretty", "cigar", "artist",
		"rough", "broken", "gorilla", "desk", "curtain", "draw",
		"bracket", "pumpkin", "family", "faculty", "snack", "laptop",
	}

	mnemonic := strings.Join(testMnemonics, " ")
	seed := bip39.NewSeed(mnemonic, "")

	master, ch := hd.ComputeMastersFromSeed(seed)
	priv, err := hd.DerivePrivateKeyForPath(master, ch, "m/86'/0'/0'/0/0")
	require.NoError(t, err)

	privKey := &taproot.PrivKey{Key: priv}
	pub := privKey.PubKey()

	accAddr := sdk.AccAddress(pub.Address()).String()

	// Bitcoin wallet generated address.
	require.Equal(t, "bc1pyxl5d7d6ejk7k3ufgnlpd4yrq939kexq4zl9yz47jw756lcjd28qa4gzrz", accAddr)
}
