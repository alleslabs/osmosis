package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Hooks is a type alias for an array of Hook.
type Hooks []Hook

// Hook is an interface of a hook that can be processed along with the ABCI application
type Hook interface {
	AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain)
	AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock)
	AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx)
	AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock)

	BeforeCommit()
}

// AfterInitChain gets called when initializing a new chain. Usually as a last step in app.InitChainer.
func (h Hooks) AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, res abci.ResponseInitChain) {
	for _, hook := range h {
		hook.AfterInitChain(ctx, req, res)
	}
}

// AfterBeginBlock gets called before completing a BeginBlocker.
func (h Hooks) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	for _, hook := range h {
		hook.AfterBeginBlock(ctx, req, res)
	}
}

// AfterDeliverTx gets called after each transaction in a block completes DeliverTx process.
func (h Hooks) AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	for _, hook := range h {
		hook.AfterDeliverTx(ctx, req, res)
	}
}

// AfterEndBlock gets called before completing an EndBlocker.
func (h Hooks) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) {
	for _, hook := range h {
		hook.AfterEndBlock(ctx, req, res)
	}
}

// BeforeCommit gets called right before committing each block.
func (h Hooks) BeforeCommit() {
	for _, hook := range h {
		hook.BeforeCommit()
	}
}
