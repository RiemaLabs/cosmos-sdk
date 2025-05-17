package bech32

import (
	"fmt"
	"runtime"

	"github.com/cosmos/btcutil/bech32"
)

// ConvertAndEncode converts from a base256 encoded byte string to base32 encoded byte string and then to bech32.
func ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("encoding bech32 failed: %w", err)
	}

	return bech32.Encode(hrp, converted)
}

// DecodeAndConvert decodes a bech32 encoded string and converts to base256 encoded bytes.
func DecodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.Decode(bech, 1023)
	if err != nil {
		// @nubit: We print the stack trace to debug the issue.
		// Remove this before the release.
		buf := make([]byte, 10240)
		runtime.Stack(buf, false)
		fmt.Println("stack trace:", string(buf))
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}

	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		// @nubit: We print the stack trace to debug the issue.
		// Remove this before the release.
		buf := make([]byte, 10240)
		runtime.Stack(buf, false)
		fmt.Println("stack trace:", string(buf))
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}

	return hrp, converted, nil
}
