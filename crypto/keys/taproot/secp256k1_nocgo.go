package taproot

import (
	"encoding/hex"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Sign creates a BIP-322 simple signature.
// It obeys the behavior of the wallet: utf-8 bytes of a string message is signed.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	privKeyObj := secp256k1.PrivKeyFromBytes(privKey.Key)
	msgHex := hex.EncodeToString(msg)
	signedMsg := []byte(msgHex)
	return Bip322Sign(signedMsg, privKeyObj, BitcoinNetParams)
}

// VerifyBytes verifies a signature of the BIP-322 simple signature.
// It obeys the behavior of the wallet: utf-8 bytes of a string message is verified.
func (pubKey *PubKey) VerifySignature(msg, sigStr []byte) bool {
	msgHex := hex.EncodeToString(msg)
	signedMsg := []byte(msgHex)
	res, err := Bip322Verify(signedMsg, sigStr, pubKey, BitcoinNetParams)
	if err != nil {
		return false
	}
	return res
}
