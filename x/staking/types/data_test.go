package types_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/taproot"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pk1      = taproot.GenPrivKey().PubKey()
	pk1Any   *codectypes.Any
	pk2      = taproot.GenPrivKey().PubKey()
	pk3      = taproot.GenPrivKey().PubKey()
	valAddr1 = sdk.ValAddress(pk1.Address())
	valAddr2 = sdk.ValAddress(pk2.Address())
	valAddr3 = sdk.ValAddress(pk3.Address())
)

func init() {
	var err error
	pk1Any, err = codectypes.NewAnyWithValue(pk1)
	if err != nil {
		panic(fmt.Sprintf("Can't pack pk1 %t as Any", pk1))
	}
}
