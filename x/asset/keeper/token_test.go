package keeper_test

import (
	"github.com/realiotech/realio-network/x/asset/types"
)

func (suite *KeeperTestSuite) TestSetToken() {
	suite.SetupTest()

	token := types.Token{
		Name:        "token",
		Symbol:      "Test",
		Decimal:     6,
		Description: "Test",
	}

	suite.app.AssetKeeper.SetToken(suite.ctx, "tokenId", token)

	expectedToken, found := suite.app.AssetKeeper.GetToken(suite.ctx, "tokenId")

	suite.Require().True(found)
	suite.Require().Equal(token, expectedToken)
}

// func (suite *KeeperTestSuite) TestSetTokenManagement() {
// 	suite.SetupTest()

// 	tokenManagement := types.TokenManagement{
// 		Manager:            suite.testUser1Address,
// 		AddNewPrivilege:    true,
// 		ExcludedPrivileges: []string{},
// 	}

// 	suite.app.AssetKeeper.SetTokenManagement(suite.ctx, "tokenId", tokenManagement)

// 	expectedTokenManagement, found := suite.app.AssetKeeper.GetTokenManagement(suite.ctx, "tokenId")

// 	suite.Require().True(found)
// 	suite.Require().Equal(tokenManagement, expectedTokenManagement)
// }

// func (suite *KeeperTestSuite) TestSetTokenPrivilegedAccount() {
// 	suite.SetupTest()

// 	suite.app.AssetKeeper.SetTokenPrivilegedAccount(suite.ctx, "tokenId", "mint", suite.testUser1Acc)

// 	privilegeAccount, found := suite.app.AssetKeeper.GetTokenPrivilegedAccount(suite.ctx, "tokenId", "mint")

// 	suite.Require().True(found)
// 	suite.Require().Equal(suite.testUser1Acc, privilegeAccount)
// }
