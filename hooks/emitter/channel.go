package emitter

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	feetypes "github.com/cosmos/ibc-go/v4/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v19/app/params"
	"github.com/osmosis-labs/osmosis/v19/hooks/common"
)

var (
	_ Adapter = &IBCAdapter{}
)

// IBCAdapter defines a struct containing the necessary flag to process the IBC hook. It implements Adapter interface.
type IBCAdapter struct {
	isIBC bool
}

// NewIBCAdapter creates a new IBCAdapter instance that will be added to the emitter hook adapters.
func NewIBCAdapter() *IBCAdapter {
	return &IBCAdapter{isIBC: false}
}

// AfterInitChain does nothing since no action is required in the InitChainer.
func (ibca *IBCAdapter) AfterInitChain(_ sdk.Context, _ params.EncodingConfig, _ map[string]json.RawMessage, _ *[]common.Message) {
}

// AfterBeginBlock does nothing since no action is required in the BeginBlocker.
func (ibca *IBCAdapter) AfterBeginBlock(_ sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
}

// PreDeliverTx sets the necessary flag to the default value before processing each transaction.
func (ibca *IBCAdapter) PreDeliverTx() {
	ibca.isIBC = false
}

// CheckMsg checks the message type and sets the IBCAdapter flag accordingly.
func (ibca *IBCAdapter) CheckMsg(_ sdk.Context, msg sdk.Msg) {
	switch msg.(type) {
	case *ibcchanneltypes.MsgRecvPacket:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelOpenInit:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelOpenTry:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelOpenAck:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelOpenConfirm:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgTimeout:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelCloseInit:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgChannelCloseConfirm:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgTimeoutOnClose:
		ibca.isIBC = true
	case *ibcchanneltypes.MsgAcknowledgement:
		ibca.isIBC = true
	case *ibcclienttypes.MsgCreateClient:
		ibca.isIBC = true
	case *ibcclienttypes.MsgUpdateClient:
		ibca.isIBC = true
	case *ibcclienttypes.MsgUpgradeClient:
		ibca.isIBC = true
	case *ibcclienttypes.MsgSubmitMisbehaviour:
		ibca.isIBC = true
	case *ibcconnectiontypes.MsgConnectionOpenInit:
		ibca.isIBC = true
	case *ibcconnectiontypes.MsgConnectionOpenTry:
		ibca.isIBC = true
	case *ibcconnectiontypes.MsgConnectionOpenAck:
		ibca.isIBC = true
	case *ibcconnectiontypes.MsgConnectionOpenConfirm:
		ibca.isIBC = true
	case *feetypes.MsgRegisterPayee:
		ibca.isIBC = true
	case *feetypes.MsgRegisterCounterpartyPayee:
		ibca.isIBC = true
	case *feetypes.MsgPayPacketFee:
		ibca.isIBC = true
	case *feetypes.MsgPayPacketFeeAsync:
		ibca.isIBC = true
	case *transfertypes.MsgTransfer:
		ibca.isIBC = true
	}
}

// HandleMsgEvents classifies IBC transaction events and assigns the values to the IBCAdapter flag accordingly.
func (ibca *IBCAdapter) HandleMsgEvents(_ sdk.Context, _ []byte, _ sdk.Msg, evMap common.EvMap, _ common.JsDict, _ *[]common.Message) {
	if _, ok := evMap[ibcchanneltypes.EventTypeSendPacket+"."+ibcchanneltypes.AttributeKeyData]; ok {
		ibca.isIBC = true
	}
}

// PostDeliverTx assigns IBCAdapter flag values to the interface to be written to the message queue.
func (ibca *IBCAdapter) PostDeliverTx(_ sdk.Context, _ []byte, txDict common.JsDict, _ *[]common.Message) {
	txDict["is_ibc"] = ibca.isIBC
}

// AfterEndBlock does nothing since no action is required in the EndBlocker.
func (ibca *IBCAdapter) AfterEndBlock(_ sdk.Context, _ abci.RequestEndBlock, _ common.EvMap, _ *[]common.Message) {
}
