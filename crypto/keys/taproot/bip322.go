package taproot

import (
	fmt "fmt"

	"github.com/babylonlabs-io/babylon/crypto/bip322"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func Bip322Verify(
	msg []byte,
	signature []byte,
	pubKey *PubKey,
	net *chaincfg.Params) (bool, error) {

	address, err := TweakedPubKeyToP2trAddress(pubKey.Address().Bytes(), net)
	if err != nil {
		return false, err
	}

	witness, err := bip322.SimpleSigToWitness(signature)
	if err != nil {
		return false, err
	}

	err = bip322.Verify(msg, witness, address, net)
	if err != nil {
		fmt.Println("error", err.Error())
		return false, err
	}

	return true, nil
}

func Bip322Sign(msg []byte, privKey *secp256k1.PrivateKey, net *chaincfg.Params) ([]byte, error) {
	emptySignature := []byte{}
	address, err := PubKeyToP2trAddress(privKey.PubKey(), net)
	if err != nil {
		return emptySignature, err
	}

	pkScript, err := TweakedPubKeyToTaprootScript(address.WitnessProgram())
	if err != nil {
		return emptySignature, err
	}

	toSpend, err := bip322.GetToSpendTx(msg, address)
	if err != nil {
		return emptySignature, err
	}

	toSign := bip322.GetToSignTx(toSpend)

	prevFetcher := txscript.NewCannedPrevOutputFetcher(
		pkScript, 0,
	)
	sigHashes := txscript.NewTxSigHashes(toSign, prevFetcher)

	witness, err := txscript.TaprootWitnessSignature(
		toSign, sigHashes, 0, 0, pkScript,
		txscript.SigHashDefault, privKey,
	)
	if err != nil {
		return emptySignature, err
	}

	signature, err := bip322.SerializeWitness(witness)
	if err != nil {
		return emptySignature, err
	}

	return signature, nil
}
