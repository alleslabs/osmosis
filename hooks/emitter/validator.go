package emitter

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v17/app/params"
	"github.com/osmosis-labs/osmosis/v17/hooks/common"
)

var _ Adapter = &ValidatorAdapter{}

// ValidatorAdapter defines a struct containing the required keeper to process the validator related hook.
// It implements Adapter interface.
type ValidatorAdapter struct {
	keeper keeper.Keeper
}

// NewValidatorAdapter creates a new ValidatorAdapter instance that will be added to the emitter hook adapters.
func NewValidatorAdapter(keeper *keeper.Keeper) *ValidatorAdapter {
	return &ValidatorAdapter{
		keeper: *keeper,
	}
}

// AfterInitChain extracts validators from the given genesis state.
func (va *ValidatorAdapter) AfterInitChain(ctx sdk.Context, encodingConfig params.EncodingConfig, genesisState map[string]json.RawMessage, kafka *[]common.Message) {
	var genutilState genutiltypes.GenesisState
	encodingConfig.Marshaler.MustUnmarshalJSON(genesisState[genutiltypes.ModuleName], &genutilState)
	for _, genTx := range genutilState.GenTxs {
		tx, err := encodingConfig.TxConfig.TxJSONDecoder()(genTx)
		if err != nil {
			panic(err)
		}
		for _, msg := range tx.GetMsgs() {
			if msg, ok := msg.(*stakingtypes.MsgCreateValidator); ok {
				valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
				va.emitSetValidator(ctx, valAddr, kafka)
			}
		}
	}
}

// AfterBeginBlock emits a new block message and handles validator jailing events.
func (va *ValidatorAdapter) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, evMap common.EvMap, kafka *[]common.Message) {
	validator, _ := va.keeper.GetValidatorByConsAddr(ctx, req.Header.GetProposerAddress())
	common.AppendMessage(kafka, "NEW_BLOCK", common.JsDict{
		"height":    req.Header.GetHeight(),
		"timestamp": ctx.BlockTime().UnixNano(),
		"proposer":  validator.GetOperator().String(),
		"hash":      req.GetHash(),
	})
	va.handleJailedEvents(ctx, evMap, kafka)
}

// PreDeliverTx does nothing since no action is required before processing each transaction.
func (va *ValidatorAdapter) PreDeliverTx() {
}

// CheckMsg does nothing since no message check is required for staking module.
func (va *ValidatorAdapter) CheckMsg(_ sdk.Context, _ sdk.Msg) {
}

// HandleMsgEvents checks for a successful message that might require validator info updating and processes them
// correspondingly.
func (va *ValidatorAdapter) HandleMsgEvents(ctx sdk.Context, _ []byte, msg sdk.Msg, _ common.EvMap, detail common.JsDict, kafka *[]common.Message) {
	switch msg := msg.(type) {
	case *stakingtypes.MsgCreateValidator:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
		val := va.emitSetValidator(ctx, valAddr, kafka)
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	case *stakingtypes.MsgEditValidator:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
		val := va.emitSetValidator(ctx, valAddr, kafka)
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	case *stakingtypes.MsgDelegate:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
		val := va.emitSetValidator(ctx, valAddr, kafka)
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	case *stakingtypes.MsgUndelegate:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
		val := va.emitSetValidator(ctx, valAddr, kafka)
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	case *stakingtypes.MsgBeginRedelegate:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
		val, _ := va.keeper.GetValidator(ctx, valAddr)
		detail["src_moniker"] = val.Description.Moniker
		detail["src_identity"] = val.Description.Identity
		valAddr, _ = sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
		val, _ = va.keeper.GetValidator(ctx, valAddr)
		detail["dst_moniker"] = val.Description.Moniker
		detail["dst_identity"] = val.Description.Identity
	case *slashingtypes.MsgUnjail:
		valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddr)
		val := va.emitSetValidator(ctx, valAddr, kafka)
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	default:
		return
	}
}

// PostDeliverTx does nothing since no action is required after the transaction has been processed by the hook.
func (va *ValidatorAdapter) PostDeliverTx(_ sdk.Context, _ []byte, _ common.JsDict, _ *[]common.Message) {
}

// AfterEndBlock only handles validator jailing events in a similar fashion to the AfterBeginBlock.
func (va *ValidatorAdapter) AfterEndBlock(ctx sdk.Context, _ abci.RequestEndBlock, evMap common.EvMap, kafka *[]common.Message) {
	va.handleJailedEvents(ctx, evMap, kafka)
}

// emitSetValidator appends the latest validator information into the provided Kafka messages array.
func (va *ValidatorAdapter) emitSetValidator(ctx sdk.Context, addr sdk.ValAddress, kafka *[]common.Message) stakingtypes.Validator {
	val, _ := va.keeper.GetValidator(ctx, addr)
	pub, _ := val.ConsPubKey()
	common.AppendMessage(kafka, "SET_VALIDATOR", common.JsDict{
		"operator_address":      addr.String(),
		"delegator_address":     sdk.AccAddress(addr).String(),
		"consensus_address":     sdk.GetConsAddress(pub).String(),
		"moniker":               val.Description.Moniker,
		"identity":              val.Description.Identity,
		"website":               val.Description.Website,
		"details":               val.Description.Details,
		"commission_rate":       val.Commission.Rate.String(),
		"commission_max_rate":   val.Commission.MaxRate.String(),
		"commission_max_change": val.Commission.MaxChangeRate.String(),
		"min_self_delegation":   val.MinSelfDelegation.String(),
		"jailed":                val.Jailed,
	})
	return val
}

// handleJailedEvents checks for slashing events and update the slashed validator accordingly.
func (va *ValidatorAdapter) handleJailedEvents(ctx sdk.Context, evMap common.EvMap, kafka *[]common.Message) {
	if raws, ok := evMap[slashingtypes.EventTypeSlash+"."+slashingtypes.AttributeKeyJailed]; ok {
		for _, raw := range raws {
			consAddress, _ := sdk.ConsAddressFromBech32(raw)
			validator, _ := va.keeper.GetValidatorByConsAddr(ctx, consAddress)
			common.AppendMessage(kafka, "UPDATE_VALIDATOR", common.JsDict{
				"operator_address": validator.OperatorAddress,
				"jailed":           validator.Jailed,
			})
		}
	}
}
