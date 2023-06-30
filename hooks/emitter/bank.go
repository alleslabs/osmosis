package emitter

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v15/app/params"
	"github.com/osmosis-labs/osmosis/v15/hooks/common"
)

var (
	_ Adapter = &BankAdapter{}
)

// BankAdapter defines a struct containing necessary flags to process the x/bank hook. It implements Adapter interface.
type BankAdapter struct {
	isSendTx bool
}

// NewBankAdapter creates new BankAdapter instance that will be added to the emitter hook adapters.
func NewBankAdapter() *BankAdapter {
	return &BankAdapter{isSendTx: false}
}

// AfterInitChain does nothing since no action is required in the InitChainer.
func (ba *BankAdapter) AfterInitChain(_ sdk.Context, _ params.EncodingConfig, _ map[string]json.RawMessage, _ *[]common.Message) {
}

// AfterBeginBlock does nothing since no action is required in the BeginBlocker.
func (ba *BankAdapter) AfterBeginBlock(_ sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
}

// PreDeliverTx sets the necessary flag to the default value before processing each transaction.
func (ba *BankAdapter) PreDeliverTx() {
	ba.isSendTx = false
}

// CheckMsg checks the message type and sets the BankAdapter flag accordingly.
func (ba *BankAdapter) CheckMsg(_ sdk.Context, msg sdk.Msg) {
	switch msg.(type) {
	case *banktypes.MsgSend:
		ba.isSendTx = true
	case *banktypes.MsgMultiSend:
		ba.isSendTx = true
	default:
		return
	}
}

// HandleMsgEvents does nothing since no action is required in the transaction events-handling step.
func (ba *BankAdapter) HandleMsgEvents(_ sdk.Context, _ []byte, _ sdk.Msg, _ common.EvMap, _ common.JsDict, _ *[]common.Message) {
}

// PostDeliverTx assigns BankAdapter flag values to the interface to be written to the message queue.
func (ba *BankAdapter) PostDeliverTx(_ sdk.Context, _ []byte, txDict common.JsDict, _ *[]common.Message) {
	txDict["is_send"] = ba.isSendTx
}

// AfterEndBlock does nothing since no action is required in the EndBlocker.
func (ba *BankAdapter) AfterEndBlock(_ sdk.Context, _ abci.RequestEndBlock, _ common.EvMap, _ *[]common.Message) {
}
