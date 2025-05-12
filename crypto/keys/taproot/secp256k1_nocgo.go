package taproot

import (
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Sign creates a BIP-322 simple signature.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	privKeyObj := secp256k1.PrivKeyFromBytes(privKey.Key)
	return Bip322Sign(msg, privKeyObj, BitcoinNetParams)
}

// VerifyBytes verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func (pubKey *PubKey) VerifySignature(msg, sigStr []byte) bool {
	res, err := Bip322Verify(msg, sigStr, pubKey, BitcoinNetParams)
	if err != nil {
		return false
	}
	return res
}
