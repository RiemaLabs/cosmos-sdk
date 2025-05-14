package codec

import (
	"errors"
	"strings"

	"cosmossdk.io/core/address"

	originBtcutil "github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type taprootCodec struct {
	btcNetworkParams *chaincfg.Params
}

var _ address.Codec = &taprootCodec{}

func NewTaprootCodec(btcNetworkParams *chaincfg.Params) address.Codec {
	return taprootCodec{btcNetworkParams}
}

// StringToBytes encodes text to bytes
func (bc taprootCodec) StringToBytes(text string) ([]byte, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	address, err := originBtcutil.DecodeAddress(text, bc.btcNetworkParams)
	if err != nil {
		return nil, err
	}

	if _, ok := address.(*originBtcutil.AddressTaproot); !ok {
		return nil, errors.New("address is not a taproot address")
	}

	if address.ScriptAddress() == nil {
		return nil, errors.New("invalid taproot address")
	}

	return address.ScriptAddress(), nil
}

// BytesToString decodes bytes to text
func (bc taprootCodec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}

	if len(bz) != 32 {
		return "", errors.New("invalid taproot address")
	}

	addr, err := originBtcutil.NewAddressTaproot(bz, bc.btcNetworkParams)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}
