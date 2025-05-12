package taproot

import (
	fmt "fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func TestSign(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("test")
	net := &chaincfg.MainNetParams
	signature, err := Bip322Sign(msg, privKey, net)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("signature", signature)
}

func TestSignAndVerify(t *testing.T) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("test")
	net := &chaincfg.MainNetParams
	signature, err := Bip322Sign(msg, privKey, net)
	if err != nil {
		t.Fatal(err)
	}
	cosmosPrivKey := PrivKey{
		Key: privKey.Serialize(),
	}
	pubKey := cosmosPrivKey.PubKey()
	verified, err := Bip322Verify(msg, signature, pubKey.(*PubKey), net)
	if err != nil {
		fmt.Println("error", err)
		t.Fatal(err)
	}
	fmt.Println(verified)
}
