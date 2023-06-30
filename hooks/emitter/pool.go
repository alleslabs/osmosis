package emitter

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/osmosis-labs/osmosis/v15/app/params"
	"github.com/osmosis-labs/osmosis/v15/hooks/common"
	gammkeeper "github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v15/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

var _ Adapter = &PoolAdapter{}

// PoolAdapter defines a struct containing required keepers, maps and flags to process the Osmosis pools related hook.
// It implements Adapter interface.
type PoolAdapter struct {
	// tx flags
	isSwapTx       bool
	isLpTx         bool
	isBondTx       bool
	isSuperfluidTx bool

	// pool transactions
	poolTxs       map[uint64]bool
	poolInBlock   map[uint64]bool
	onlyPoolMsgTx bool

	// keepers
	gammKeeper   *gammkeeper.Keeper
	govKeeper    *govkeeper.Keeper
	lockupKeeper *lockupkeeper.Keeper
}

// NewPoolAdapter creates a new PoolAdapter instance that will be added to the emitter hook adapters.
func NewPoolAdapter(
	gammKeeper *gammkeeper.Keeper,
	govKeeper *govkeeper.Keeper,
	lockupKeeper *lockupkeeper.Keeper,
) *PoolAdapter {
	return &PoolAdapter{
		isSwapTx:       false,
		isLpTx:         false,
		isBondTx:       false,
		isSuperfluidTx: false,
		poolTxs:        make(map[uint64]bool),
		poolInBlock:    make(map[uint64]bool),
		onlyPoolMsgTx:  true,
		gammKeeper:     gammKeeper,
		govKeeper:      govKeeper,
		lockupKeeper:   lockupKeeper,
	}
}

// AfterInitChain does nothing since no action is required in the InitChainer.
func (pa *PoolAdapter) AfterInitChain(
	_ sdk.Context,
	_ params.EncodingConfig,
	_ map[string]json.RawMessage,
	_ *[]common.Message,
) {
}

// AfterBeginBlock sets the necessary map to the starting value before processing each block.
func (pa *PoolAdapter) AfterBeginBlock(_ sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
	pa.poolInBlock = make(map[uint64]bool)
}

// PreDeliverTx sets the necessary maps and flags to the starting value before processing each transaction.
func (pa *PoolAdapter) PreDeliverTx() {
	pa.isSwapTx = false
	pa.isLpTx = false
	pa.isBondTx = false
	pa.isSuperfluidTx = false
	pa.poolTxs = make(map[uint64]bool)
	pa.onlyPoolMsgTx = true
}

// CheckMsg checks the message type and extracts message values to PoolAdapter maps and flags accordingly.
func (pa *PoolAdapter) CheckMsg(ctx sdk.Context, msg sdk.Msg) {
	switch msg := msg.(type) {
	case *gammtypes.MsgJoinPool:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *gammtypes.MsgExitPool:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *gammtypes.MsgSwapExactAmountIn:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			pa.poolTxs[route.PoolId] = true
		}
	case *gammtypes.MsgSwapExactAmountOut:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			pa.poolTxs[route.PoolId] = true
		}
	case *gammtypes.MsgJoinSwapExternAmountIn:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *gammtypes.MsgJoinSwapShareAmountOut:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *gammtypes.MsgExitSwapExternAmountOut:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *gammtypes.MsgExitSwapShareAmountIn:
		pa.isLpTx = true
		pa.poolTxs[msg.PoolId] = true
	case *poolmanagertypes.MsgSwapExactAmountIn:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			pa.poolTxs[route.PoolId] = true
		}
	case *poolmanagertypes.MsgSwapExactAmountOut:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			pa.poolTxs[route.PoolId] = true
		}
	case *lockuptypes.MsgLockTokens:
		if poolIds, found := getPoolIdsFromCoins(msg.Coins); found {
			pa.isBondTx = true
			for _, poolId := range poolIds {
				pa.poolTxs[poolId] = true
			}
		}
	case *lockuptypes.MsgForceUnlock:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.ID); found {
			pa.isBondTx = true
			pa.poolTxs[poolId] = true
		}
	case *lockuptypes.MsgBeginUnlocking:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.ID); found {
			pa.isBondTx = true
			pa.poolTxs[poolId] = true
		}
	case *lockuptypes.MsgExtendLockup:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.ID); found {
			pa.isBondTx = true
			pa.poolTxs[poolId] = true
		}
	case *superfluidtypes.MsgSuperfluidDelegate:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.LockId); found {
			pa.isSuperfluidTx = true
			pa.poolTxs[poolId] = true
		}
	case *superfluidtypes.MsgSuperfluidUndelegate:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.LockId); found {
			pa.isSuperfluidTx = true
			pa.poolTxs[poolId] = true
		}
	case *superfluidtypes.MsgSuperfluidUnbondLock:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.LockId); found {
			pa.isSuperfluidTx = true
			pa.poolTxs[poolId] = true
		}
	case *superfluidtypes.MsgLockAndSuperfluidDelegate:
		if poolIds, found := getPoolIdsFromCoins(msg.Coins); found {
			pa.isSuperfluidTx = true
			for _, poolId := range poolIds {
				pa.poolTxs[poolId] = true
			}
		}
	case *superfluidtypes.MsgUnPoolWhitelistedPool:
		pa.isSuperfluidTx = true
		pa.poolTxs[msg.PoolId] = true
	case *superfluidtypes.MsgSuperfluidUndelegateAndUnbondLock:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.LockId); found {
			pa.isSuperfluidTx = true
			pa.poolTxs[poolId] = true
		}
	default:
		pa.onlyPoolMsgTx = false
	}
}

// HandleMsgEvents processes a special case which could not be handled during CheckMsg because it has to be a success
// transaction only. Also, handles pool events emitted from messages not from Osmosis pools related modules, e.g.,
// via contract execution.
func (pa *PoolAdapter) HandleMsgEvents(
	ctx sdk.Context,
	txHash []byte,
	msg sdk.Msg,
	evMap common.EvMap,
	_ common.JsDict,
	kafka *[]common.Message,
) {
	switch msg := msg.(type) {
	case *lockuptypes.MsgBeginUnlockingAll:
		pa.handleMsgBeginUnlockingAll(ctx, evMap)
	default:
		if !pa.onlyPoolMsgTx {
			pa.handleCreatePoolEvents(ctx, txHash, msg.GetSigners()[0], evMap, kafka)
			pa.handleNonPoolMsgsPoolActionEvents(ctx, evMap)
		}
	}
}

// PostDeliverTx appends the processed transaction to the array of messages to be written to Kafka and adds the pool
// to the array of pools to be updated at end block.
func (pa *PoolAdapter) PostDeliverTx(ctx sdk.Context, txHash []byte, _ common.JsDict, kafka *[]common.Message) {
	for poolId := range pa.poolTxs {
		common.AppendMessage(kafka, "NEW_POOL_TRANSACTION", common.JsDict{
			"pool_id":       poolId,
			"tx_hash":       txHash,
			"block_height":  ctx.BlockHeight(),
			"is_swap":       pa.isSwapTx,
			"is_lp":         pa.isLpTx,
			"is_bond":       pa.isBondTx,
			"is_superfluid": pa.isSuperfluidTx,
		})
		pa.poolInBlock[poolId] = true
	}
	pa.poolTxs = make(map[uint64]bool)
}

// AfterEndBlock checks for superfluid related ActiveProposal events and update the status accordingly. Update pool
// stats at the end.
func (pa *PoolAdapter) AfterEndBlock(
	ctx sdk.Context,
	_ abci.RequestEndBlock,
	evMap common.EvMap,
	kafka *[]common.Message,
) {
	if rawIds, ok := evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			proposalId := uint64(common.Atoi(rawId))
			proposal, _ := pa.govKeeper.GetProposal(ctx, proposalId)

			switch content := proposal.GetContent().(type) {
			case *superfluidtypes.SetSuperfluidAssetsProposal:
				for _, asset := range content.Assets {
					if id, ok := getPoolIdFromDenom(asset.Denom); ok {
						common.AppendMessage(kafka, "UPDATE_SET_SUPERFLUID_ASSET", common.JsDict{
							"id": common.Atoi(id),
						})
					}
				}
			case *superfluidtypes.RemoveSuperfluidAssetsProposal:
				for _, denom := range content.SuperfluidAssetDenoms {
					if id, ok := getPoolIdFromDenom(denom); ok {
						common.AppendMessage(kafka, "UPDATE_REMOVE_SUPERFLUID_ASSET", common.JsDict{
							"id": common.Atoi(id),
						})
					}
				}
			}
		}
	}
	pa.flushUpdatePoolStats(ctx, kafka)
}

// getPoolIdFromDenom returns pool id if the provided denom is a pool share token. Otherwise, returns false.
func getPoolIdFromDenom(denom string) (string, bool) {
	if !strings.HasPrefix(denom, "gamm/pool/") {
		return "", false
	}
	return strings.Trim(denom, "gamm/pool/"), true
}

// getPoolIdsFromCoins filters an array of coins and returns pool ids of pool share tokens in the array.
func getPoolIdsFromCoins(coins sdk.Coins) ([]uint64, bool) {
	result := make([]uint64, 0)
	found := false
	for _, coin := range coins {
		if poolId, ok := getPoolIdFromDenom(coin.GetDenom()); ok {
			result = append(result, common.Atoui(poolId))
			found = true
		}
	}

	return result, found
}

// getPoolIdFromLockId returns pool id of the given lock if the locked token is a pool share token.
func (pa *PoolAdapter) getPoolIdFromLockId(ctx sdk.Context, lockId uint64) (uint64, bool) {
	lock, err := pa.lockupKeeper.GetLockByID(ctx, lockId)
	if err != nil {
		return 0, false
	}
	for _, coin := range lock.Coins {
		if poolId, ok := getPoolIdFromDenom(coin.GetDenom()); ok {
			return common.Atoui(poolId), true
		}
	}
	return 0, false
}

// getPoolIdsFromLockIds filters an array of locks and returns pool ids of pool share tokens in the locks.
func (pa *PoolAdapter) getPoolIdsFromLockIds(ctx sdk.Context, lockIds []string) []uint64 {
	result := make([]uint64, 0)
	for _, lockId := range lockIds {
		lock, err := pa.lockupKeeper.GetLockByID(ctx, common.Atoui(lockId))
		if err != nil {
			continue
		}
		for _, coin := range lock.Coins {
			if poolId, ok := getPoolIdFromDenom(coin.GetDenom()); ok {
				result = append(result, common.Atoui(poolId))
			}
		}
	}

	return result
}

// handleCreatePoolEvents handles PoolCreated events and processes new pools in the current transaction.
func (pa *PoolAdapter) handleCreatePoolEvents(ctx sdk.Context, txHash []byte, sender sdk.AccAddress, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[gammtypes.TypeEvtPoolCreated+"."+gammtypes.AttributeKeyPoolId]; ok {
		newPool := make(map[uint64]bool)
		for _, rawId := range rawIds {
			newPool[common.Atoui(rawId)] = true
		}
		for poolId := range newPool {
			poolInfo, _ := pa.gammKeeper.GetPool(ctx, poolId)
			switch pool := poolInfo.(type) {
			case *balancer.Pool:
				weights := make([]common.JsDict, 0)
				for _, weight := range pool.PoolAssets {
					weights = append(weights, common.JsDict{
						"denom":  weight.Token.GetDenom(),
						"weight": weight.Weight.String(),
					})
				}
				poolMsg := common.JsDict{
					"id":                   poolId,
					"liquidity":            pool.GetTotalPoolLiquidity(ctx),
					"type":                 pool.GetType(),
					"creator":              sender.String(),
					"create_tx":            txHash,
					"is_superfluid":        false,
					"is_supported":         false,
					"swap_fee":             pool.GetSwapFee(ctx).String(),
					"exit_fee":             pool.GetExitFee(ctx).String(),
					"future_pool_governor": pool.FuturePoolGovernor,
					"weight":               weights,
					"address":              pool.GetAddress().String(),
					"total_shares":         pool.TotalShares,
				}
				if smoothWeightChangeParams := pool.PoolParams.GetSmoothWeightChangeParams(); smoothWeightChangeParams != nil {
					poolMsg["smooth_weight_change_params"] = smoothWeightChangeParams
				}
				common.AppendMessage(kafka, "NEW_GAMM_POOL", poolMsg)
			case *stableswap.Pool:
				common.AppendMessage(kafka, "NEW_GAMM_POOL", common.JsDict{
					"id":                        poolId,
					"liquidity":                 pool.GetTotalPoolLiquidity(ctx),
					"type":                      pool.GetType(),
					"creator":                   sender.String(),
					"create_tx":                 txHash,
					"is_superfluid":             false,
					"is_supported":              false,
					"swap_fee":                  pool.GetSwapFee(ctx).String(),
					"exit_fee":                  pool.GetExitFee(ctx).String(),
					"future_pool_governor":      pool.FuturePoolGovernor,
					"scaling_factors":           pool.GetScalingFactors(),
					"scaling_factor_controller": pool.ScalingFactorController,
					"address":                   pool.GetAddress().String(),
					"total_shares":              pool.TotalShares,
				})
			default:
				panic("cannot handle pool type")
			}
		}
	}
}

// handleNonPoolMsgsPoolActionEvents handles events emitted from calling other messages calling pool actions
// (from contracts, ica)
func (pa *PoolAdapter) handleNonPoolMsgsPoolActionEvents(ctx sdk.Context, evMap common.EvMap) {
	if poolIds, ok := evMap[gammtypes.TypeEvtPoolJoined+"."+gammtypes.AttributeKeyPoolId]; ok {
		pa.isLpTx = true
		for _, poolId := range poolIds {
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if poolIds, ok := evMap[gammtypes.TypeEvtPoolExited+"."+gammtypes.AttributeKeyPoolId]; ok {
		pa.isLpTx = true
		for _, poolId := range poolIds {
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if poolIds, ok := evMap[gammtypes.TypeEvtTokenSwapped+"."+gammtypes.AttributeKeyPoolId]; ok {
		pa.isSwapTx = true
		for _, poolId := range poolIds {
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if lockIds, ok := evMap[lockuptypes.TypeEvtLockTokens+"."+lockuptypes.AttributePeriodLockID]; ok {
		pa.isBondTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[lockuptypes.TypeEvtAddTokensToLock+"."+lockuptypes.AttributePeriodLockID]; ok {
		pa.isBondTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[lockuptypes.TypeEvtBeginUnlock+"."+lockuptypes.AttributePeriodLockID]; ok {
		pa.isBondTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[superfluidtypes.TypeEvtSuperfluidDelegate+"."+superfluidtypes.AttributeLockId]; ok {
		pa.isSuperfluidTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[superfluidtypes.TypeEvtSuperfluidUndelegate+"."+superfluidtypes.AttributeLockId]; ok {
		pa.isSuperfluidTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[superfluidtypes.TypeEvtSuperfluidUnbondLock+"."+superfluidtypes.AttributeLockId]; ok {
		pa.isSuperfluidTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
	if lockIds, ok := evMap[superfluidtypes.TypeEvtSuperfluidUndelegateAndUnbondLock+"."+superfluidtypes.AttributeLockId]; ok {
		pa.isSuperfluidTx = true
		poolIds := pa.getPoolIdsFromLockIds(ctx, lockIds)
		for _, poolId := range poolIds {
			pa.poolTxs[poolId] = true
		}
	}
}

// handleMsgBeginUnlockingAll gets all pool ids from lock ids being unlocking via the message.
func (pa *PoolAdapter) handleMsgBeginUnlockingAll(ctx sdk.Context, evMap common.EvMap) {
	pa.isBondTx = true
	if lockIds, ok := evMap[lockuptypes.TypeEvtBeginUnlock+"."+lockuptypes.AttributePeriodLockID]; ok {
		for _, poolId := range pa.getPoolIdsFromLockIds(ctx, lockIds) {
			pa.poolTxs[poolId] = true
		}
	}
}

// flushUpdatePoolStats appends updated Osmosis pools stats into the provided Kafka messages array.
func (pa *PoolAdapter) flushUpdatePoolStats(ctx sdk.Context, kafka *[]common.Message) {
	for poolId := range pa.poolInBlock {
		poolInfo, _ := pa.gammKeeper.GetPool(ctx, poolId)
		switch pool := poolInfo.(type) {
		case *balancer.Pool:
			weights := make([]common.JsDict, 0)
			for _, weight := range pool.PoolAssets {
				weights = append(weights, common.JsDict{
					"denom":  weight.Token.GetDenom(),
					"weight": weight.Weight.String(),
				})
			}
			common.AppendMessage(kafka, "UPDATE_POOL", common.JsDict{
				"id":           poolId,
				"liquidity":    pool.GetTotalPoolLiquidity(ctx),
				"weight":       weights,
				"total_shares": pool.TotalShares,
			})
		case *stableswap.Pool:
			common.AppendMessage(kafka, "UPDATE_POOL", common.JsDict{
				"id":           poolId,
				"liquidity":    pool.GetTotalPoolLiquidity(ctx),
				"total_shares": pool.TotalShares,
			})
		}
	}
	pa.poolInBlock = make(map[uint64]bool)
}
