package emitter

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v17/app/params"
	"github.com/osmosis-labs/osmosis/v17/hooks/common"
	"github.com/osmosis-labs/osmosis/v17/x/protorev/keeper"
)

var (
	_ Adapter = &ProtorevAdapter{}
)

// ProtorevAdapter defines a struct containing the required keeper to process the Osmosis x/protorev hook.
// It implements Adapter interface.
type ProtorevAdapter struct {
	keeper *keeper.Keeper
}

// NewProtorevAdapter creates a new ProtorevAdapter instance that will be added to the emitter hook adapters.
func NewProtorevAdapter(keeper *keeper.Keeper) *ProtorevAdapter {
	return &ProtorevAdapter{
		keeper: keeper,
	}
}

// AfterInitChain does nothing since no action is required in the InitChainer.
func (pa *ProtorevAdapter) AfterInitChain(_ sdk.Context, _ params.EncodingConfig, _ map[string]json.RawMessage, _ *[]common.Message) {
}

// AfterBeginBlock does nothing since no action is required in the BeginBlocker.
func (pa *ProtorevAdapter) AfterBeginBlock(_ sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
}

// PreDeliverTx does nothing since no action is required before processing each transaction.
func (pa *ProtorevAdapter) PreDeliverTx() {
}

// CheckMsg does nothing since no message check is required for x/protorev module.
func (pa *ProtorevAdapter) CheckMsg(_ sdk.Context, _ sdk.Msg) {
}

// HandleMsgEvents does nothing since no action is required in the transaction events-handling step.
func (pa *ProtorevAdapter) HandleMsgEvents(_ sdk.Context, _ []byte, _ sdk.Msg, _ common.EvMap, _ common.JsDict, _ *[]common.Message) {
}

// PostDeliverTx does nothing since no action is required after the transaction has been processed by the hook.
func (pa *ProtorevAdapter) PostDeliverTx(_ sdk.Context, _ []byte, _ common.JsDict, _ *[]common.Message) {
}

// AfterEndBlock appends new x/protorev data for the current block into the provided Kafka messages array.
func (pa *ProtorevAdapter) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, _ common.EvMap, kafka *[]common.Message) {
	trade, _ := pa.keeper.GetNumberOfTrades(ctx)
	common.AppendMessage(kafka, "TRADE", common.JsDict{
		"block_height": req.Height,
		"count":        trade,
	})

	for _, profit := range pa.keeper.GetAllProfits(ctx) {
		common.AppendMessage(kafka, "PROFIT_BY_DENOM", common.JsDict{
			"block_height": req.Height,
			"denom":        profit.Denom,
			"amount":       profit.Amount,
		})
	}

	routes, _ := pa.keeper.GetAllRoutes(ctx)
	for _, route := range routes {
		tradeByRoute, _ := pa.keeper.GetTradesByRoute(ctx, route)
		common.AppendMessage(kafka, "TRADE_BY_ROUTE", common.JsDict{
			"block_height": req.Height,
			"route":        route,
			"count":        tradeByRoute,
		})
		profits := pa.keeper.GetAllProfitsByRoute(ctx, route)
		for _, profit := range profits {
			common.AppendMessage(kafka, "PROFIT_BY_ROUTE", common.JsDict{
				"block_height": req.Height,
				"route":        route,
				"denom":        profit.Denom,
				"amount":       profit.Amount,
			})
		}
	}
}
