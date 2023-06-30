package emitter

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/osmosis-labs/osmosis/v15/hooks/common"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/rpc/core/types"
)

// getTxResponse returns an unmarshalled Cosmos SDK TxResponse from Tendermint RequestDeliverTx and ResponseDeliverTx
func (h *Hook) getTxResponse(ctx sdk.Context, txHash []byte, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) common.JsDict {
	resTx := tmtypes.ResultTx{
		Hash:     txHash,
		Height:   ctx.BlockHeight(),
		TxResult: res,
		Tx:       req.Tx,
	}
	txResult, err := common.MkTxResult(h.encodingConfig.TxConfig, &resTx, ctx.BlockTime())
	if err != nil {
		panic(err)
	}
	protoTx, ok := txResult.Tx.GetCachedValue().(*txtypes.Tx)
	if !ok {
		panic("cannot make proto tx")
	}
	txResJson, err := codec.ProtoMarshalJSON(&txtypes.GetTxResponse{
		Tx:         protoTx,
		TxResponse: txResult,
	}, nil)
	if err != nil {
		panic(err)
	}
	var txResJsDict common.JsDict
	err = json.Unmarshal(txResJson, &txResJsDict)
	if err != nil {
		panic(err)
	}
	return txResJsDict
}

// getMessageDicts returns an array of JsDict decoded version for messages in the provided transaction.
func getMessageDicts(txResJsDict common.JsDict) []common.JsDict {
	details := make([]common.JsDict, 0)
	tx := txResJsDict["tx"].(map[string]interface{})
	body := tx["body"].(map[string]interface{})
	msgs := body["messages"].([]interface{})
	for _, msg := range msgs {
		detail := msg.(map[string]interface{})
		details = append(details, detail)
	}
	return details
}

// getTxDict returns a JsDict decoded version for the provided transaction.
func getTxDict(ctx sdk.Context, tx sdk.Tx, txHash []byte, res abci.ResponseDeliverTx) common.JsDict {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		panic(fmt.Sprintf("cannot get fee tx for tx %s", txHash))
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		panic(fmt.Sprintf("cannot get memo for tx %s", txHash))
	}
	var errMsg *string
	if !res.IsOK() {
		errMsg = &res.Log
	}

	return common.JsDict{
		"hash":         txHash,
		"block_height": ctx.BlockHeight(),
		"gas_used":     res.GasUsed,
		"gas_limit":    feeTx.GetGas(),
		"gas_fee":      feeTx.GetFee().String(),
		"err_msg":      errMsg,
		"sender":       tx.GetMsgs()[0].GetSigners()[0].String(),
		"success":      res.IsOK(),
		"memo":         memoTx.GetMemo(),
	}
}

// updateTxInBlockAndRelatedTx is being called at the end of each DeliverTx to update hook accounts in transaction
// and accounts in block maps.
func (h *Hook) updateTxInBlockAndRelatedTx(ctx sdk.Context, txHash []byte, signers []string) {
	h.AddAccountsInTx(signers...)
	relatedAccounts := make([]string, 0)
	for acc := range h.accsInTx {
		relatedAccounts = append(relatedAccounts, acc)
	}
	h.AddAccountsInBlock(relatedAccounts...)
	common.AppendMessage(&h.msgs, "SET_RELATED_TRANSACTION", common.JsDict{
		"hash":             txHash,
		"block_height":     ctx.BlockHeight(),
		"signer":           signers,
		"related_accounts": relatedAccounts,
	})
}
