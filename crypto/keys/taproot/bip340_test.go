package taproot

import (
	"encoding/hex"
	fmt "fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/require"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Keep the same with Rust.
// let private_key = PrivateKey::new(SecretKey::from_slice(&hex::decode(secret).unwrap()).unwrap(), network);
// let public_key = PublicKey::from_private_key(&secp, &private_key);
// let address = generate_p2tr_address(network, public_key);

func TestKeySpendPathP2trAddress(t *testing.T) {

	keyDatas := []keyData{
		{
			priv: "a6f1b630d3a9cefdd2f5f81a5e647dd9e8ed89330c65186adfa455388a8a1bc5",
			pub:  "02a46430ecf90d8207825f2b1063c52a53c3ddbb5d8d77060f3b5a88386dc97e01",
			addr: "bc1pnz4wc9q7f0qlcc2uxty233ahhdt5rhue879ularqrzvj8y3wmq0s9mysgu",
		},
		{
			priv: "d67c3041fbdba5c9b6a9f400c8207769671025d7eb847f69bf690867ddc63160",
			pub:  "034590d3ab8746bd9a28cd2c658832e94e34100a3cb2785d5efc8db3bc23697ee9",
			addr: "bc1pn0whuyajs800qprsh6cltfd5f4xg98waq70pea8vtd860m4kydtqk4uxta",
		},
		{
			priv: "af0e23f1b2a749dd99c4b2191f55bb1b516a5a16ae807b47ae4a5573f0cdeb73",
			pub:  "0274b8812cbf44b4556bb6e998653cabb2174e9cb95ba1a8dbbeb2fdc417db048c",
			addr: "bc1pgg65gl9n6xg33nyurpmfpnswgqeamzthwclyuufmr7grumrrdvcqjlm2gx",
		},
		{
			priv: "6148c7cec662b4161a960321c1a92e706bddd70dc05f3b46bde0031f9bc2b046",
			pub:  "02a12f14bad1e2285fca2a0bd37e21b7d6b7eaec815a9aa92b679aaf26a28408f0",
			addr: "bc1puumagkr448llkx665whqec3mm3ve2l28ykvj4m0d8amry5ussh5sj0flay",
		},
	}

	for _, keyData := range keyDatas {
		privKeyBytes, err := hex.DecodeString(keyData.priv)
		require.NoError(t, err)
		privKeyObj := secp256k1.PrivKeyFromBytes(privKeyBytes)
		pubkeyObject := privKeyObj.PubKey()

		// Original pubkey checking.
		pubKeyBytes, err := hex.DecodeString(keyData.pub)
		require.NoError(t, err)
		require.Equal(t, pubKeyBytes, pubkeyObject.SerializeCompressed())

		// Address checking.
		address, err := PubKeyToP2trAddress(pubkeyObject, &chaincfg.MainNetParams)
		require.NoError(t, err)
		require.Equal(t, address.String(), keyData.addr)

		// Tweaked pubkey checking.
		tweakedPubKey := TweakPubKey(pubkeyObject)
		require.Equal(t, tweakedPubKey[1:], address.ScriptAddress())

		// Tweaked pubkey to address checking.
		tweakedPubKeyBytes := tweakedPubKey[1:]
		address, err = TweakedPubKeyToP2trAddress(tweakedPubKeyBytes, &chaincfg.MainNetParams)
		require.NoError(t, err)
		require.Equal(t, address.String(), keyData.addr)

		// Print out the witness program.
		fmt.Println(hex.EncodeToString(address.WitnessProgram()))
	}
}
