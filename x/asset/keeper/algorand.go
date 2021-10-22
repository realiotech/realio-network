package keeper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/realiotech/network/x/asset/types"
	"net"
	"strings"
)

// Executes send to algorand transaction and returns in case errors
// If error occurs no token will be minted or deleted...

func ExecuteSendToAlgorand(goCtx context.Context, msg *types.MsgSendToAlgorand, assetKeeper Keeper) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// check balance of the account ....
	// try parsing the payment to coins to see if user has sufficient balance
	var payment sdk.Coin
	if msg.Denom == "rio" {
		payment, err = sdk.ParseCoinNormalized(fmt.Sprintf("%v%s", msg.Amount, "rio"))
		if err != nil {
			return errors.New("Uable to parse coins...")
		}
	} else {
		return errors.New("Invalid denom...")
	}

	// check if they have sufficient balance
	hasBalance := assetKeeper.bankKeeper.HasBalance(ctx, sender, payment)
	if !hasBalance {
		return errors.New("Your balance is less than given amount...")
	}

	// TODO: Signature

	// TODO: BURN THE TOKENS HERE ...
	// This is similar to locking the token
	// In case the method returns error then
	// the tokens will be minted back to your account???
	err = assetKeeper.BurnTokens(ctx, sender, payment)
	// this shouldn't be happening...
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent("unable to burn tokens", sdk.NewAttribute("failed", "burn")))
		return errors.New("Unable to safe burn tokens")
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent("Burned existing tokens...", sdk.NewAttribute("amount", msg.Amount)))


	instruction := fmt.Sprintf("mint %s %v\n", msg.AlgorandReceiver, msg.Amount)


	conn, err := net.Dial("tcp", types.RelayerAddr)
	if err != nil {
		return err
	}

	fmt.Fprintf(conn, instruction)

	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}

	if strings.Contains(strings.TrimSpace(string(status)), "FAIL") {
		return errors.New(string(status))
	} else if strings.Contains(strings.TrimSpace(string(status)), "SUCCESS") {
		// if minted or unlocked tokens in tezos chain successfully
		// pass the message to message server so that it can burn the amount, denom of coins in cosmos pegzone...
		// do nothing because tokens are already burned
		return nil

	}

	return errors.New("look like there is still an error...")
}