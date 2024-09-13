package mint

type MintPriviledge struct {
	bk BankKeeper
}

func NewMintPriviledge(bk BankKeeper) MintPriviledge {
	return MintPriviledge{
		bk: bk,
	}
}
