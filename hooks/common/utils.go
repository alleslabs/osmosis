package common

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	cmtypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// EvMap is a type alias for SDK events mapping from Attr.Key to the list of values.
type EvMap map[string][]string

// JsDict is a type alias for JSON dictionary.
type JsDict map[string]interface{}

// Message is a simple wrapper data type for each message published to Kafka.
type Message struct {
	Key   string
	Value JsDict
}

// Atoi converts the given string into an int64. Panics on errors.
func Atoi(val string) int64 {
	res, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

// Atoui converts the given string into an uint64. Panics on errors.
func Atoui(val string) uint64 {
	res, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

// ParseEvents parses the given sdk.StringEvents objects into a single EvMap.
func ParseEvents(events sdk.StringEvents) EvMap {
	evMap := make(EvMap)
	for _, event := range events {
		for _, kv := range event.Attributes {
			key := event.Type + "." + kv.Key
			evMap[key] = append(evMap[key], kv.Value)
		}
	}
	return evMap
}

// BuildContractAddressClassic builds an SDK account address for a CosmWasm contract.
func BuildContractAddressClassic(codeID, instanceID uint64) sdk.AccAddress {
	contractID := make([]byte, 16)
	binary.BigEndian.PutUint64(contractID[:8], codeID)
	binary.BigEndian.PutUint64(contractID[8:], instanceID)
	return address.Module(types.ModuleName, contractID)[:types.ContractAddrLen]
}

// GetRelatedAccounts finds all valid sdk.AccAddress recursively in the given interface.
func GetRelatedAccounts(m interface{}, accs map[string]bool) {
	switch m := m.(type) {
	case []interface{}:
		for _, v := range m {
			GetRelatedAccounts(v, accs)
		}
	case map[string]interface{}:
		for _, v := range m {
			GetRelatedAccounts(v, accs)
		}
	case string:
		_, err := sdk.AccAddressFromBech32(m)
		if err == nil {
			accs[m] = true
		}
	default:
	}
}

// Deprecated: this interface is used only internally for some scenarios we are
// deprecating (StdTxConfig support)
type intoAny interface {
	AsAny() *codectypes.Any
}

// MkTxResult returns a sdk.TxResponse from the given Tendermint ResultTx.
func MkTxResult(txConfig client.TxConfig, resTx *cmtypes.ResultTx, blockTime time.Time) (*sdk.TxResponse, error) {
	txb, err := txConfig.TxDecoder()(resTx.Tx)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	asAny := p.AsAny()
	return sdk.NewResponseResultTx(resTx, asAny, blockTime.Format(time.RFC3339)), nil
}

// AppendMessage is a simple function for appending new key and value to an array of Message.
func AppendMessage(msgs *[]Message, key string, value JsDict) {
	*msgs = append(*msgs, Message{Key: key, Value: value})
}
