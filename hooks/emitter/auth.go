package emitter

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	"github.com/osmosis-labs/osmosis/v16/hooks/common"
)

// AccountType enumerates the valid types of an account.
type AccountType int32

const (
	BaseAccount              AccountType = 0
	InterchainAccount        AccountType = 1
	ModuleAccount            AccountType = 2
	ContinuousVestingAccount AccountType = 3
	DelayedVestingAccount    AccountType = 4
	ClawbackVestingAccount   AccountType = 5
	ContractAccount          AccountType = 6
)

var (
	_ AccountVerifier = &ContractAccountVerifier{}
	_ AccountVerifier = &AuthAccountVerifier{}
)

// AccountVerifier defines an interface for the implementation of verifying a valid account.
type AccountVerifier interface {
	VerifyAccount(ctx sdk.Context, addr sdk.AccAddress) (common.JsDict, bool)
}

// ContractAccountVerifier is an AccountVerifier implementation for verifying CosmWasm contract accounts.
type ContractAccountVerifier struct {
	keeper wasmkeeper.Keeper
}

// VerifyAccount returns the account and its type if the given input is a valid CosmWasm contract account.
func (ca ContractAccountVerifier) VerifyAccount(ctx sdk.Context, addr sdk.AccAddress) (common.JsDict, bool) {
	if c := ca.keeper.GetContractInfo(ctx, addr); c != nil {
		return common.JsDict{
			"address": addr.String(),
			"type":    ContractAccount,
		}, true
	}
	return nil, false
}

// AuthAccountVerifier is an AccountVerifier implementation for verifying x/auth and ica accounts.
type AuthAccountVerifier struct {
	keeper authkeeper.AccountKeeper
}

// VerifyAccount returns the account and its type if the given input is a valid x/auth or ica account.
func (aa AuthAccountVerifier) VerifyAccount(ctx sdk.Context, addr sdk.AccAddress) (common.JsDict, bool) {
	acc := aa.keeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, false
	}

	if moduleAccount, ok := acc.(types.ModuleAccountI); ok {
		return common.JsDict{
			"address": addr.String(),
			"type":    ModuleAccount,
			"name":    moduleAccount.GetName(),
		}, true
	}
	if _, ok := acc.(*icatypes.InterchainAccount); ok {
		return common.JsDict{
			"address": addr.String(),
			"type":    InterchainAccount,
		}, true
	}
	if _, ok := acc.(*vestingtypes.ContinuousVestingAccount); ok {
		return common.JsDict{
			"address": addr.String(),
			"type":    ContinuousVestingAccount,
		}, true
	}
	if _, ok := acc.(*vestingtypes.DelayedVestingAccount); ok {
		return common.JsDict{
			"address": addr.String(),
			"type":    DelayedVestingAccount,
		}, true
	}
	if _, ok := acc.(*vestingtypes.ClawbackVestingAccount); ok {
		return common.JsDict{
			"address": addr.String(),
			"type":    ClawbackVestingAccount,
		}, true
	}
	return common.JsDict{
		"address": addr.String(),
		"type":    BaseAccount,
	}, true
}

// VerifyAccount iterates through each given AccountVerifier and returns its address and account type according to the input address.
func VerifyAccount(ctx sdk.Context, addr sdk.AccAddress, avs ...AccountVerifier) common.JsDict {
	for _, verifier := range avs {
		if d, ok := verifier.VerifyAccount(ctx, addr); ok {
			return d
		}
	}
	return common.JsDict{
		"address": addr.String(),
	}
}
