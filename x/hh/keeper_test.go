package hh

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

var ModuleBasics module.BasicManager

func init() {

	ModuleBasics = module.NewBasicManager(
		AppModule{},
	)
}

func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func makeAcc() sdk.AccAddress {
	var pub ed25519.PubKeyEd25519
	rand.Read(pub[:])
	return sdk.AccAddress(pub.Address())

}

func TestPutTwoNFTOnMarket(t *testing.T) {
	stKey := sdk.NewKVStoreKey(StoreKey)
	ti := setupTestInput(stKey)
	k := NewKeeper(nil, stKey, ti.cdc)

	account := makeAcc()
	someToken := NFT{
		BaseNFT{
			ID: "1234",
		},
		false,
		nil,
	}
	price := sdk.Coins{sdk.Coin{
		"usd",
		sdk.NewInt(100),
	}}
	k.setNFTOwner(ti.ctx, someToken.BaseNFT.ID, account)
	//put first NFT
	err := k.PutNFTokenOnTheMarket(ti.ctx, someToken.BaseNFT, price, account)
	if err != nil {
		t.Fatal(err)
	}
	nftList := k.GetNFTokensOnSaleList(ti.ctx)
	if len(nftList) != 1 {
		t.Fatal("incorrect length")
	}

	newToken := someToken
	newToken.ID = newToken.ID + "1"
	k.setNFTOwner(ti.ctx, newToken.ID, account)
	err = k.PutNFTokenOnTheMarket(ti.ctx, newToken.BaseNFT, sdk.Coins{sdk.Coin{
		"usd",
		sdk.NewInt(150),
	}}, account)
	if err != nil {
		t.Fatal(err)
	}
	nftList = k.GetNFTokensOnSaleList(ti.ctx)
	if len(nftList) != 2 {
		t.Fatal("incorrect length")
	}
}

func TestPutSameNFTOnMarket(t *testing.T) {
	stKey := sdk.NewKVStoreKey(StoreKey)
	ti := setupTestInput(stKey)
	k := NewKeeper(nil, stKey, ti.cdc)
	someToken := NFT{
		BaseNFT{
			ID: "1234",
		},
		false,
		nil,
	}

	price := sdk.Coins{sdk.Coin{
		"usd",
		sdk.NewInt(100),
	}}
	account := makeAcc()
	k.setNFTOwner(ti.ctx, someToken.ID, account)
	//put first NFT
	err := k.PutNFTokenOnTheMarket(ti.ctx, someToken.BaseNFT, price, account)
	if err != nil {
		t.Fatal(err)
	}

	nftList := k.GetNFTokensOnSaleList(ti.ctx)
	if len(nftList) != 1 {
		t.Fatal("incorrect length")
	}

	err = k.PutNFTokenOnTheMarket(ti.ctx, someToken.BaseNFT, sdk.Coins{sdk.Coin{
		"usd",
		sdk.NewInt(150),
	}}, sdk.AccAddress{})
	if err == nil {
		t.FailNow()
	}
	nftList = k.GetNFTokensOnSaleList(ti.ctx)
	if len(nftList) != 1 {
		t.Fatal("incorrect length")
	}

}

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
}

func setupTestInput(key sdk.StoreKey) testInput {
	db := dbm.NewMemDB()

	cdc := MakeCodec()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	return testInput{cdc: cdc, ctx: ctx}
}
