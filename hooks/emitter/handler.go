package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/hooks/common"
)

// handleMsg finds related accounts for the given message events and handles the message in each adapter in the hook.
func (h *Hook) handleMsg(ctx sdk.Context, txHash []byte, msg sdk.Msg, log sdk.ABCIMessageLog, detail common.JsDict) {
	evMap := common.ParseEvents(log.Events)
	for _, values := range evMap {
		for _, value := range values {
			if _, err := sdk.AccAddressFromBech32(value); err == nil {
				h.AddAccountsInTx(value)
			}
		}
	}

	for idx := range h.adapters {
		h.adapters[idx].HandleMsgEvents(ctx, txHash, msg, evMap, detail, &h.msgs)
	}
}
