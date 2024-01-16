package emitter

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v22/app/params"
	"github.com/osmosis-labs/osmosis/v22/hooks/common"
)

// Adapter defines an interface of an adapter for each emitter-supported module to be processed along with the emitter hook.
type Adapter interface {
	AfterInitChain(ctx sdk.Context, encodingConfig params.EncodingConfig, genesisState map[string]json.RawMessage, kafka *[]common.Message)
	AfterBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, evMap common.EvMap, kafka *[]common.Message)
	PreDeliverTx()
	CheckMsg(ctx sdk.Context, msg sdk.Msg)
	HandleMsgEvents(ctx sdk.Context, txHash []byte, msg sdk.Msg, evMap common.EvMap, detail common.JsDict, kafka *[]common.Message)
	PostDeliverTx(ctx sdk.Context, txHash []byte, txDict common.JsDict, kafka *[]common.Message)
	AfterEndBlock(ctx sdk.Context, req abci.RequestEndBlock, evMap common.EvMap, kafka *[]common.Message)
}
