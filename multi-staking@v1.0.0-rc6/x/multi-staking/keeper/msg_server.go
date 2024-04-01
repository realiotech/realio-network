package keeper

import (
	"context"
	"fmt"

	"github.com/realio-tech/multi-staking-module/x/multi-staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type msgServer struct {
	keeper           Keeper
	stakingMsgServer stakingtypes.MsgServer
}

var _ stakingtypes.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) stakingtypes.MsgServer {
	return &msgServer{
		keeper:           keeper,
		stakingMsgServer: stakingkeeper.NewMsgServerImpl(keeper.stakingKeeper),
	}
}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(goCtx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	multiStakerAddr, valAcc, err := types.AccAddrAndValAddrFromStrings(msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	lockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorAddress)

	mintedBondCoin, err := k.keeper.LockCoinAndMintBondCoin(ctx, lockID, multiStakerAddr, multiStakerAddr, msg.Value)
	if err != nil {
		return nil, err
	}

	sdkMsg := stakingtypes.MsgCreateValidator{
		Description:       msg.Description,
		Commission:        msg.Commission,
		MinSelfDelegation: msg.MinSelfDelegation,
		DelegatorAddress:  msg.DelegatorAddress,
		ValidatorAddress:  msg.ValidatorAddress,
		Pubkey:            msg.Pubkey,
		Value:             mintedBondCoin, // replace lock coin with bond coin
	}

	k.keeper.SetValidatorMultiStakingCoin(ctx, valAcc, msg.Value.Denom)

	return k.stakingMsgServer.CreateValidator(sdk.WrapSDKContext(ctx), &sdkMsg)
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(goCtx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return k.stakingMsgServer.EditValidator(goCtx, msg)
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(goCtx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	multiStakerAddr, valAcc, err := types.AccAddrAndValAddrFromStrings(msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if !k.keeper.isValMultiStakingCoin(ctx, valAcc, msg.Amount) {
		return nil, fmt.Errorf("not allowed coin")
	}

	lockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorAddress)

	mintedBondCoin, err := k.keeper.LockCoinAndMintBondCoin(ctx, lockID, multiStakerAddr, multiStakerAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	sdkMsg := stakingtypes.MsgDelegate{
		DelegatorAddress: msg.DelegatorAddress,
		ValidatorAddress: msg.ValidatorAddress,
		Amount:           mintedBondCoin, // replace lock coin with bond coin
	}

	return k.stakingMsgServer.Delegate(sdk.WrapSDKContext(ctx), &sdkMsg)
}

// BeginRedelegate defines a method for performing a redelegation of coins from a delegator and source validator to a destination validator
func (k msgServer) BeginRedelegate(goCtx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	multiStakerAddr := sdk.MustAccAddressFromBech32(msg.DelegatorAddress)

	srcValAcc, err := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	if err != nil {
		return nil, err
	}
	dstValAcc, err := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	if err != nil {
		return nil, err
	}

	if !k.keeper.isValMultiStakingCoin(ctx, srcValAcc, msg.Amount) || !k.keeper.isValMultiStakingCoin(ctx, dstValAcc, msg.Amount) {
		return nil, fmt.Errorf("not allowed Coin")
	}

	fromLockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorSrcAddress)
	fromLock, found := k.keeper.GetMultiStakingLock(ctx, fromLockID)
	if !found {
		return nil, fmt.Errorf("lock not found")
	}

	toLockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorDstAddress)
	toLock := k.keeper.GetOrCreateMultiStakingLock(ctx, toLockID)

	multiStakingCoin := fromLock.MultiStakingCoin(msg.Amount.Amount)

	err = fromLock.MoveCoinToLock(&toLock, multiStakingCoin)
	if err != nil {
		return nil, err
	}
	k.keeper.SetMultiStakingLock(ctx, fromLock)
	k.keeper.SetMultiStakingLock(ctx, toLock)

	redelegateAmount := multiStakingCoin.BondValue()
	redelegateAmount, err = k.keeper.AdjustUnbondAmount(ctx, multiStakerAddr, srcValAcc, redelegateAmount)
	if err != nil {
		return nil, err
	}

	bondCoin := sdk.NewCoin(k.keeper.stakingKeeper.BondDenom(ctx), redelegateAmount)

	sdkMsg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    msg.DelegatorAddress,
		ValidatorSrcAddress: msg.ValidatorSrcAddress,
		ValidatorDstAddress: msg.ValidatorDstAddress,
		Amount:              bondCoin, // replace lockCoin with bondCoin
	}

	return k.stakingMsgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), sdkMsg)
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(goCtx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	multiStakerAddr, valAcc, err := types.AccAddrAndValAddrFromStrings(msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if !k.keeper.isValMultiStakingCoin(ctx, valAcc, msg.Amount) {
		return nil, fmt.Errorf("not allowed coin")
	}

	lockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorAddress)
	lock, found := k.keeper.GetMultiStakingLock(ctx, lockID)
	if !found {
		return nil, fmt.Errorf("can't find multi staking lock")
	}

	multiStakingCoin := lock.MultiStakingCoin(msg.Amount.Amount)
	err = lock.RemoveCoinFromMultiStakingLock(multiStakingCoin)
	if err != nil {
		return nil, err
	}
	k.keeper.SetMultiStakingLock(ctx, lock)

	unbondAmount := multiStakingCoin.BondValue()
	unbondAmount, err = k.keeper.AdjustUnbondAmount(ctx, multiStakerAddr, valAcc, unbondAmount)
	if err != nil {
		return nil, err
	}

	unbondCoin := sdk.NewCoin(k.keeper.stakingKeeper.BondDenom(ctx), unbondAmount)

	sdkMsg := &stakingtypes.MsgUndelegate{
		DelegatorAddress: msg.DelegatorAddress,
		ValidatorAddress: msg.ValidatorAddress,
		Amount:           unbondCoin, // replace with unbondCoin
	}

	k.keeper.SetMultiStakingUnlockEntry(ctx, types.MultiStakingUnlockID(msg.DelegatorAddress, msg.ValidatorAddress), multiStakingCoin)

	return k.stakingMsgServer.Undelegate(sdk.WrapSDKContext(ctx), sdkMsg)
}

// CancelUnbondingDelegation defines a method for canceling the unbonding delegation
// and delegate back to the validator.
func (k msgServer) CancelUnbondingDelegation(goCtx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delAcc, valAcc, err := types.AccAddrAndValAddrFromStrings(msg.DelegatorAddress, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	if !k.keeper.isValMultiStakingCoin(ctx, valAcc, msg.Amount) {
		return nil, fmt.Errorf("not allow coin")
	}

	unlockID := types.MultiStakingUnlockID(msg.DelegatorAddress, msg.ValidatorAddress)
	cancelUnlockingCoin, err := k.keeper.DecreaseUnlockEntryAmount(ctx, unlockID, msg.Amount.Amount, msg.CreationHeight)
	if err != nil {
		return nil, err
	}

	cancelUnbondingAmount := cancelUnlockingCoin.BondValue()
	cancelUnbondingAmount, err = k.keeper.AdjustCancelUnbondingAmount(ctx, delAcc, valAcc, msg.CreationHeight, cancelUnbondingAmount)
	if err != nil {
		return nil, err
	}
	cancelUnbondingCoin := sdk.NewCoin(k.keeper.stakingKeeper.BondDenom(ctx), cancelUnbondingAmount)

	lockID := types.MultiStakingLockID(msg.DelegatorAddress, msg.ValidatorAddress)
	lock := k.keeper.GetOrCreateMultiStakingLock(ctx, lockID)
	err = lock.AddCoinToMultiStakingLock(cancelUnlockingCoin)
	if err != nil {
		return nil, err
	}
	k.keeper.SetMultiStakingLock(ctx, lock)

	sdkMsg := &stakingtypes.MsgCancelUnbondingDelegation{
		DelegatorAddress: msg.DelegatorAddress,
		ValidatorAddress: msg.ValidatorAddress,
		Amount:           cancelUnbondingCoin, // replace with cancelUnbondingCoin
		CreationHeight:   msg.CreationHeight,
	}

	return k.stakingMsgServer.CancelUnbondingDelegation(sdk.WrapSDKContext(ctx), sdkMsg)
}
