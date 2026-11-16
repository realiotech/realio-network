package keeper_test

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/realiotech/realio-network/testutil"
	"github.com/realiotech/realio-network/x/claim"
	"github.com/realiotech/realio-network/x/claim/types"
)

func (suite *KeeperTestSuite) TestHackedAmountSetGetClear() {
	k := suite.app.ClaimKeeper
	victim, other := testutil.GenAddress(), testutil.GenAddress()

	got, err := k.GetHackedAmount(suite.ctx, victim)
	suite.Require().NoError(err)
	suite.Require().True(got.IsZero(), "nothing is recorded by default")

	amount := sdk.NewCoins(sdk.NewInt64Coin("ario", 500), sdk.NewInt64Coin("arst", 7))
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, victim, amount))
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, other, coins(9)))

	got, err = k.GetHackedAmount(suite.ctx, victim)
	suite.Require().NoError(err)
	suite.Require().Equal(amount, got, "every denom is read back, in canonical order")

	// Setting again replaces rather than merges: a denom dropped from the
	// new amount must not linger from the old one.
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, victim, coins(1)))
	got, err = k.GetHackedAmount(suite.ctx, victim)
	suite.Require().NoError(err)
	suite.Require().Equal(coins(1), got)

	// Clearing one address leaves its neighbours alone.
	suite.Require().NoError(k.ClearHackedAmount(suite.ctx, victim))
	got, err = k.GetHackedAmount(suite.ctx, victim)
	suite.Require().NoError(err)
	suite.Require().True(got.IsZero())
	got, err = k.GetHackedAmount(suite.ctx, other)
	suite.Require().NoError(err)
	suite.Require().Equal(coins(9), got)

	// An empty amount clears.
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, other, sdk.Coins{}))
	got, err = k.GetHackedAmount(suite.ctx, other)
	suite.Require().NoError(err)
	suite.Require().True(got.IsZero())
}

func (suite *KeeperTestSuite) TestSetHackedAmountRejectsInvalidCoins() {
	victim := testutil.GenAddress()
	bad := sdk.Coins{sdk.Coin{Denom: "ario", Amount: math.NewInt(-1)}}
	suite.Require().Error(suite.app.ClaimKeeper.SetHackedAmount(suite.ctx, victim, bad))
	unsorted := sdk.Coins{sdk.NewInt64Coin("bbb", 1), sdk.NewInt64Coin("aaa", 1)}
	suite.Require().Error(suite.app.ClaimKeeper.SetHackedAmount(suite.ctx, victim, unsorted))
}

func (suite *KeeperTestSuite) TestGetAllHackedAmountsGroupsByAddress() {
	k := suite.app.ClaimKeeper
	a, b := testutil.GenAddress(), testutil.GenAddress()
	multi := sdk.NewCoins(sdk.NewInt64Coin("ario", 1), sdk.NewInt64Coin("arst", 2))
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, a, multi))
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, b, coins(3)))

	all, err := k.GetAllHackedAmounts(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(all, 2, "one entry per address, not per denom")

	byAddr := map[string]sdk.Coins{}
	for _, h := range all {
		byAddr[h.Address] = h.Amount
	}
	suite.Require().Equal(multi, byAddr[a.String()])
	suite.Require().Equal(coins(3), byAddr[b.String()])
}

// TestHackedAmountsSurviveGenesisRoundTrip: what is on record is exported,
// and importing the export reproduces it. A paid record must not come back.
func (suite *KeeperTestSuite) TestHackedAmountsSurviveGenesisRoundTrip() {
	k := suite.app.ClaimKeeper
	owed, paid := testutil.GenAddress(), testutil.GenAddress()
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, owed, coins(42)))
	suite.Require().NoError(k.SetHackedAmount(suite.ctx, paid, coins(8)))
	suite.Require().NoError(k.ClearHackedAmount(suite.ctx, paid))

	mod := claim.NewAppModule(k)
	bz := mod.ExportGenesis(suite.ctx, suite.app.AppCodec())

	var exported types.GenesisState
	suite.Require().NoError(json.Unmarshal(bz, &exported))
	suite.Require().Len(exported.HackedAmounts, 1)
	suite.Require().Equal(owed.String(), exported.HackedAmounts[0].Address)
	suite.Require().Equal(coins(42), exported.HackedAmounts[0].Amount)

	// Import into a clean slate.
	suite.Require().NoError(k.ClearHackedAmount(suite.ctx, owed))
	mod.InitGenesis(suite.ctx, suite.app.AppCodec(), bz)
	got, err := k.GetHackedAmount(suite.ctx, owed)
	suite.Require().NoError(err)
	suite.Require().Equal(coins(42), got)
	got, err = k.GetHackedAmount(suite.ctx, paid)
	suite.Require().NoError(err)
	suite.Require().True(got.IsZero())
}
