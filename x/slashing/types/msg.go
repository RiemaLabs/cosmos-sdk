package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgUnjail{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgUnjail creates a new MsgUnjail instance
func NewMsgUnjail(validatorAddr string, delegatorAddr string) *MsgUnjail {
	return &MsgUnjail{
		ValidatorAddr: validatorAddr,
		DelegatorAddr: delegatorAddr,
	}
}
