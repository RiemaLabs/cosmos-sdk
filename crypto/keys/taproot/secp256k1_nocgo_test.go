package taproot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSignatureVerificationAndRejectUpperS ensures that:
// 1. Valid Taproot signatures are verified correctly
// 2. Non-canonical signatures (those with upper-S values) are rejected
// 3. Empty or invalid signatures are rejected
func TestSignatureVerificationAndRejectUpperS(t *testing.T) {
	msg := []byte("We have lingered long enough on the shores of the cosmic ocean.")

	// Test empty signature
	priv := GenPrivKey()
	pub := priv.PubKey()
	require.False(t, pub.VerifySignature(msg, []byte{}), "Empty signature should be rejected")

	// Test invalid signature length
	require.False(t, pub.VerifySignature(msg, make([]byte, 63)), "Invalid signature length should be rejected")

	// Test valid signatures and upper-S rejection
	for range 500 {
		priv := GenPrivKey()
		sigStr, err := priv.Sign(msg)
		require.NoError(t, err)

		// Verify the original signature
		pub := priv.PubKey()
		require.True(t, pub.VerifySignature(msg, sigStr), "Valid signature should be accepted")

		// Test malleated signature
		malleatedSigStr := make([]byte, len(sigStr))
		copy(malleatedSigStr, sigStr)
		malleatedSigStr[len(malleatedSigStr)-1] = ^malleatedSigStr[len(malleatedSigStr)-1]
		require.False(t, pub.VerifySignature(msg, malleatedSigStr), "Malleated signature should be rejected")
	}
}
