package emitter

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmjson "github.com/cometbft/cometbft/libs/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/segmentio/kafka-go"

	"github.com/osmosis-labs/osmosis/v21/app/keepers"
	"github.com/osmosis-labs/osmosis/v21/app/params"
	"github.com/osmosis-labs/osmosis/v21/hooks/common"
)

// Hook uses Kafka message queue and adapters functionality to act as an event producer for all events in the blockchains.
type Hook struct {
	encodingConfig params.EncodingConfig // The app encoding config
	writer         *kafka.Writer         // Main Kafka writer instance
	accsInBlock    map[string]bool       // Accounts needed to be updated at the end of the block
	accsInTx       map[string]bool       // Accounts related to the current processing transaction
	msgs           []common.Message      // The list of all Kafka messages to be published for this block
	adapters       []Adapter             // Array of adapters needed for the hook
	accVerifiers   []AccountVerifier     // Array of AccountVerifier needed for account verification
}

// NewHook creates an emitter hook instance that will be added in the Osmosis App.
func NewHook(
	encodingConfig params.EncodingConfig,
	keeper keepers.AppKeepers,
	kafkaURI string,
) *Hook {
	paths := strings.SplitN(kafkaURI, "@", 2)
	return &Hook{
		encodingConfig: encodingConfig,
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      paths[1:],
			Topic:        paths[0],
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 1 * time.Millisecond,
			BatchBytes:   512000000,
		}),
		adapters: []Adapter{
			NewValidatorAdapter(keeper.StakingKeeper),
			NewBankAdapter(),
			NewIBCAdapter(),
			NewGovAdapter(keeper.GovKeeper, keeper.StakingKeeper),
			NewWasmAdapter(keeper.WasmKeeper, keeper.GovKeeper),
			NewPoolAdapter(keeper.AccountKeeper, keeper.ConcentratedLiquidityKeeper, keeper.CosmwasmPoolKeeper, keeper.GAMMKeeper, keeper.GovKeeper, keeper.LockupKeeper, keeper.PoolManagerKeeper, keeper.WasmKeeper),
			// NewProtorevAdapter(keeper.ProtoRevKeeper),
		},
		accVerifiers: []AccountVerifier{
			ContractAccountVerifier{keeper: *keeper.WasmKeeper},
			AuthAccountVerifier{keeper: *keeper.AccountKeeper},
		},
	}
}

// AddAccountsInBlock adds the given accounts to the array of accounts to be updated at EndBlocker.
func (h *Hook) AddAccountsInBlock(accs ...string) {
	for _, acc := range accs {
		h.accsInBlock[acc] = true
	}
}

// AddAccountsInTx adds the given accounts to the array of accounts related to the current processing transaction.
func (h *Hook) AddAccountsInTx(accs ...string) {
	for _, acc := range accs {
		h.accsInTx[acc] = true
	}
}

// FlushMessages publishes all pending messages to Kafka message queue. Blocks until completion.
func (h *Hook) FlushMessages() {
	kafkaMsgs := make([]kafka.Message, len(h.msgs))
	for idx, msg := range h.msgs {
		res, _ := json.Marshal(msg.Value) // Error must always be nil.
		kafkaMsgs[idx] = kafka.Message{Key: []byte(msg.Key), Value: res}
	}
	err := h.writer.WriteMessages(context.Background(), kafkaMsgs...)
	if err != nil {
		panic(err)
	}
}

// AfterInitChain specifies actions to be done after chain initialization (app.Hook interface).
func (h *Hook) AfterInitChain(ctx sdk.Context, req abci.RequestInitChain, _ abci.ResponseInitChain) {
	var genesisState map[string]json.RawMessage
	if err := cmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	var authGenesis authtypes.GenesisState
	if genesisState[authtypes.ModuleName] != nil {
		h.encodingConfig.Marshaler.MustUnmarshalJSON(genesisState[authtypes.ModuleName], &authGenesis)
	}
	for _, account := range authGenesis.GetAccounts() {
		a, ok := account.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("expected account")
		}

		common.AppendMessage(&h.msgs, "SET_ACCOUNT", VerifyAccount(ctx, a.GetAddress(), h.accVerifiers...))
	}

	for idx := range h.adapters {
		h.adapters[idx].AfterInitChain(ctx, h.encodingConfig, genesisState, &h.msgs)
	}

	common.AppendMessage(&h.msgs, "COMMIT", common.JsDict{"height": 0})
	h.FlushMessages()
}

// AfterBeginBlock specifies actions needed to be done after each BeginBlock period (app.Hook interface)
func (h *Hook) AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	h.accsInBlock = make(map[string]bool)
	h.accsInTx = make(map[string]bool)
	h.msgs = []common.Message{}
	evMap := common.ParseEvents(sdk.StringifyEvents(res.Events))
	for idx := range h.adapters {
		h.adapters[idx].AfterBeginBlock(ctx, req, evMap, &h.msgs)
	}
}

// AfterDeliverTx specifies actions to be done after each transaction has been processed (app.Hook interface).
func (h *Hook) AfterDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	if ctx.BlockHeight() == 0 {
		return
	}

	h.accsInTx = make(map[string]bool)
	for idx := range h.adapters {
		h.adapters[idx].PreDeliverTx()
	}
	txHash := tmhash.Sum(req.Tx)
	tx, err := h.encodingConfig.TxConfig.TxDecoder()(req.Tx)
	if err != nil {
		panic("cannot decode tx")
	}
	txDict := getTxDict(ctx, tx, txHash, res)
	common.AppendMessage(&h.msgs, "NEW_TRANSACTION", txDict)

	txRes := h.getTxResponse(ctx, txHash, req, res)
	common.AppendMessage(&h.msgs, "INSERT_LCD_TX_RESULTS", common.JsDict{
		"tx_hash":      txHash,
		"block_height": ctx.BlockHeight(),
		"result":       txRes,
	})
	md := getMessageDicts(txRes)
	logs, _ := sdk.ParseABCILogs(res.Log)
	var msgs []map[string]interface{}
	for idx, msg := range tx.GetMsgs() {
		for i := range h.adapters {
			h.adapters[i].CheckMsg(ctx, msg)
		}
		common.GetRelatedAccounts(h.GetMsgJson(msg), h.accsInTx)
		if res.IsOK() {
			h.handleMsg(ctx, txHash, msg, logs[idx], md[idx])
		}
		msgs = append(msgs, common.JsDict{
			"detail": md[idx],
			"type":   sdk.MsgTypeURL(msg),
		})
	}

	signers := tx.GetMsgs()[0].GetSigners()
	addrs := make([]string, len(signers))
	for idx, signer := range signers {
		addrs[idx] = signer.String()
	}

	h.updateTxInBlockAndRelatedTx(ctx, txHash, addrs)
	h.PostDeliverTx(ctx, txHash, txDict, msgs)
}

// PostDeliverTx specifies actions to be done by adapters after each transaction has been processed by the hook.
func (h *Hook) PostDeliverTx(ctx sdk.Context, txHash []byte, txDict common.JsDict, msgs []map[string]interface{}) {
	txDict["messages"] = msgs
	for idx := range h.adapters {
		h.adapters[idx].PostDeliverTx(ctx, txHash, txDict, &h.msgs)
	}
}

// AfterEndBlock specifies actions to be done after each end block period (app.Hook interface).
func (h *Hook) AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) {
	evMap := common.ParseEvents(sdk.StringifyEvents(res.Events))
	for idx := range h.adapters {
		h.adapters[idx].AfterEndBlock(ctx, req, evMap, &h.msgs)
	}

	// Index 0 is the message NEW_BLOCK, SET_ACCOUNT messages are inserted between NEW_BLOCK and other messages.
	modifiedMsgs := []common.Message{h.msgs[0]}
	for accStr := range h.accsInBlock {
		acc, _ := sdk.AccAddressFromBech32(accStr)
		modifiedMsgs = append(modifiedMsgs, common.Message{
			Key:   "SET_ACCOUNT",
			Value: VerifyAccount(ctx, acc, h.accVerifiers...),
		})
	}
	h.msgs = append(modifiedMsgs, h.msgs[1:]...)
	common.AppendMessage(&h.msgs, "COMMIT", common.JsDict{"height": req.Height})
}

// BeforeCommit specifies actions to be done before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {
	h.FlushMessages()
}

// GetMsgJson returns an unmarshalled interface of the given sdk.Msg.
func (h *Hook) GetMsgJson(msg sdk.Msg) interface{} {
	bz, _ := h.encodingConfig.Marshaler.MarshalInterfaceJSON(msg)
	var jsonMsg interface{}
	err := json.Unmarshal(bz, &jsonMsg)
	if err != nil {
		panic(err)
	}
	return jsonMsg
}
