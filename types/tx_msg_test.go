package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type testMsgSuite struct {
	suite.Suite
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(testMsgSuite))
}

func (s *testMsgSuite) TestMsg() {
	addr := []byte{0x6c, 0x6a, 0x7f, 0x10, 0xe0, 0x67, 0xe, 0xd5, 0x6f, 0x1a, 0x4a, 0xf2, 0xc, 0x8a, 0xcb, 0xf6, 0xf4, 0x8a, 0x35, 0xb2, 0xe0, 0x5d, 0x96, 0x1d, 0xf6, 0x6b, 0x18, 0x2a, 0xd, 0xba, 0xf6, 0xad}
	accAddr := sdk.AccAddress(addr)

	msg := testdata.NewTestMsg(accAddr)
	s.Require().NotNil(msg)
	s.Require().True(accAddr.Equals(msg.GetSigners()[0]))
	s.Require().Nil(msg.ValidateBasic())
}

func (s *testMsgSuite) TestMsgTypeURL() {
	s.Require().Equal("/testpb.TestMsg", sdk.MsgTypeURL(new(testdata.TestMsg)))
	s.Require().Equal("/google.protobuf.Any", sdk.MsgTypeURL(&anypb.Any{}))
}

func (s *testMsgSuite) TestGetMsgFromTypeURL() {
	msg := new(testdata.TestMsg)
	cdc := codec.NewProtoCodec(testdata.NewTestInterfaceRegistry())

	result, err := sdk.GetMsgFromTypeURL(cdc, "/testpb.TestMsg")
	s.Require().NoError(err)
	s.Require().Equal(msg, result)
}
