package writer

import (
	"context"

	"github.com/osmosis-labs/osmosis/v16/hooks/common"
)

type Writer interface {
	WriteMessages(ctx context.Context, msgs []common.Message)
}
