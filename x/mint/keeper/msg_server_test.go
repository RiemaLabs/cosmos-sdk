package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (s *IntegrationTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "set invalid authority (not an address)",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
		},
		{
			name: "set invalid authority (not defined authority)",
			request: &types.MsgUpdateParams{
				Authority: "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0",
			},
			expectErr: true,
		},
		{
			name: "set invalid params",
			request: &types.MsgUpdateParams{
				Authority: s.mintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:           sdk.DefaultBondDenom,
					InflationRateChange: sdkmath.LegacyNewDecWithPrec(-13, 2),
					InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
					InflationMin:        sdkmath.LegacyNewDecWithPrec(7, 2),
					GoalBonded:          sdkmath.LegacyNewDecWithPrec(67, 2),
					BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
				},
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.mintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:           sdk.DefaultBondDenom,
					InflationRateChange: sdkmath.LegacyNewDecWithPrec(8, 2),
					InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
					InflationMin:        sdkmath.LegacyNewDecWithPrec(2, 2),
					GoalBonded:          sdkmath.LegacyNewDecWithPrec(37, 2),
					BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
