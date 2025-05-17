package taproot

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func PubKeyToP2trAddress(p *secp256k1.PublicKey, net *chaincfg.Params) (*btcutil.AddressTaproot, error) {
	tapKey := txscript.ComputeTaprootKeyNoScript(p)

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey), net,
	)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func TweakedPubKeyToP2trAddress(p []byte, net *chaincfg.Params) (*btcutil.AddressTaproot, error) {
	if len(p) != 32 {
		return nil, errors.New("invalid pubkey length")
	}
	address, err := btcutil.NewAddressTaproot(
		p, net,
	)
	if err != nil {
		return nil, err
	}
	return address, nil
}

func TweakPubKey(p *secp256k1.PublicKey) []byte {
	tapKey := txscript.ComputeTaprootKeyNoScript(p)

	xCoor := schnorr.SerializePubKey(tapKey)

	pubKey := make([]byte, 33)
	pubKey[0] = 0x02
	copy(pubKey[1:], xCoor)

	return pubKey
}

func TweakedPubKeyToTaprootScript(p []byte) ([]byte, error) {
	if len(p) != 32 {
		return nil, errors.New("invalid pubkey length")
	}

	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_1).
		AddData(p).
		Script()
}
