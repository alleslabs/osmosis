package emitter

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v16/app/params"
	"github.com/osmosis-labs/osmosis/v16/hooks/common"
	clpkeeper "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity"
	concentratedpool "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/model"
	clptypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	cosmwasmpoolkeeper "github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool"
	cosmwasmpool "github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/model"
	cosmwasmpooltypes "github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
	gammkeeper "github.com/osmosis-labs/osmosis/v16/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	gammtypes "github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v16/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	poolmanagerkeeper "github.com/osmosis-labs/osmosis/v16/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v16/x/superfluid/types"
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
	isClp          bool
	isCollect      bool
	isMigrate      bool

	// pool transactions
	cosmwasmContractAddrs map[string]uint64
	isCosmwasmLoaded      bool
	poolTxs               map[uint64]bool
	poolInBlock           map[uint64]bool
	onlyPoolMsgTx         bool

	// keepers
	accountKeeper      *authkeeper.AccountKeeper
	clpKeeper          *clpkeeper.Keeper
	cosmwasmpoolKeeper *cosmwasmpoolkeeper.Keeper
	gammKeeper         *gammkeeper.Keeper
	govKeeper          *govkeeper.Keeper
	lockupKeeper       *lockupkeeper.Keeper
	poolmanagerKeeper  *poolmanagerkeeper.Keeper
	wasmKeeper         *wasmkeeper.Keeper
}

// NewPoolAdapter creates a new PoolAdapter instance that will be added to the emitter hook adapters.
func NewPoolAdapter(
	accountKeeper *authkeeper.AccountKeeper,
	clpKeeper *clpkeeper.Keeper,
	cosmwasmpoolKeeper *cosmwasmpoolkeeper.Keeper,
	gammKeeper *gammkeeper.Keeper,
	govKeeper *govkeeper.Keeper,
	lockupKeeper *lockupkeeper.Keeper,
	poolmanagerKeeper *poolmanagerkeeper.Keeper,
	wasmKeeper *wasmkeeper.Keeper,
) *PoolAdapter {
	return &PoolAdapter{
		isSwapTx:              false,
		isLpTx:                false,
		isBondTx:              false,
		isSuperfluidTx:        false,
		isClp:                 false,
		isCollect:             false,
		isMigrate:             false,
		cosmwasmContractAddrs: make(map[string]uint64),
		isCosmwasmLoaded:      false,
		poolTxs:               make(map[uint64]bool),
		poolInBlock:           make(map[uint64]bool),
		onlyPoolMsgTx:         true,
		accountKeeper:         accountKeeper,
		clpKeeper:             clpKeeper,
		cosmwasmpoolKeeper:    cosmwasmpoolKeeper,
		gammKeeper:            gammKeeper,
		govKeeper:             govKeeper,
		lockupKeeper:          lockupKeeper,
		poolmanagerKeeper:     poolmanagerKeeper,
		wasmKeeper:            wasmKeeper,
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
func (pa *PoolAdapter) AfterBeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
	if !pa.isCosmwasmLoaded {
		pools, err := pa.cosmwasmpoolKeeper.GetPools(ctx)
		if err != nil {
			panic(err)
		}

		for _, pool := range pools {
			pa.cosmwasmContractAddrs[pool.GetAddress().String()] = pool.GetId()
		}
		pa.isCosmwasmLoaded = true
	}
	pa.poolInBlock = make(map[uint64]bool)
}

// PreDeliverTx sets the necessary maps and flags to the starting value before processing each transaction.
func (pa *PoolAdapter) PreDeliverTx() {
	pa.isSwapTx = false
	pa.isLpTx = false
	pa.isBondTx = false
	pa.isSuperfluidTx = false
	pa.isClp = false
	pa.isCollect = false
	pa.isMigrate = false
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
	case *poolmanagertypes.MsgSplitRouteSwapExactAmountIn:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			for _, pool := range route.Pools {
				pa.poolTxs[pool.PoolId] = true
			}
		}
	case *poolmanagertypes.MsgSplitRouteSwapExactAmountOut:
		pa.isSwapTx = true
		for _, route := range msg.Routes {
			for _, pool := range route.Pools {
				pa.poolTxs[pool.PoolId] = true
			}
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
	case *lockuptypes.MsgSetRewardReceiverAddress:
		if poolId, found := pa.getPoolIdFromLockId(ctx, msg.LockID); found {
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
	case *superfluidtypes.MsgCreateFullRangePositionAndSuperfluidDelegate:
		pa.isSuperfluidTx = true
		pa.poolTxs[msg.PoolId] = true
	case *superfluidtypes.MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition:
		if poolId, found := getPoolIdFromDenom(msg.SharesToMigrate.GetDenom()); found {
			pa.isMigrate = true
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	case *superfluidtypes.MsgAddToConcentratedLiquiditySuperfluidPosition:
		if position, err := pa.clpKeeper.GetPosition(ctx, msg.PositionId); err == nil {
			pa.isSuperfluidTx = true
			pa.poolTxs[position.PoolId] = true
		}
	case *clptypes.MsgCreatePosition:
		pa.isClp = true
		pa.poolTxs[msg.PoolId] = true

	// This case works for some transaction (failed tx or partially withdraw) since the position would be deleted if
	// the whole position is withdrawn.
	case *clptypes.MsgWithdrawPosition:
		if position, err := pa.clpKeeper.GetPosition(ctx, msg.PositionId); err == nil {
			pa.isClp = true
			pa.poolTxs[position.PoolId] = true
		}
	case *clptypes.MsgAddToPosition:
		if position, err := pa.clpKeeper.GetPosition(ctx, msg.PositionId); err == nil {
			pa.isClp = true
			pa.poolTxs[position.PoolId] = true
		}
	case *clptypes.MsgCollectSpreadRewards:
		for _, positionId := range msg.PositionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, positionId); err == nil {
				pa.isCollect = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	case *clptypes.MsgCollectIncentives:
		for _, positionId := range msg.PositionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, positionId); err == nil {
				pa.isCollect = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	case *wasmtypes.MsgExecuteContract:
		if poolId, found := pa.cosmwasmContractAddrs[msg.Contract]; found {
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
	case *clptypes.MsgWithdrawPosition:
		if poolIds, ok := evMap[clptypes.TypeEvtWithdrawPosition+"."+clptypes.AttributeKeyPoolId]; ok {
			pa.isClp = true
			for _, poolId := range poolIds {
				pa.poolTxs[common.Atoui(poolId)] = true
			}
		}
	case *superfluidtypes.MsgUnlockAndMigrateSharesToFullRangeConcentratedPosition:
		// Pool id leaving is already handled in the CheckMsg via lock id, so we only need to handle pool id entering here
		if poolIds, ok := evMap[superfluidtypes.TypeEvtUnlockAndMigrateShares+"."+superfluidtypes.AttributeKeyPoolIdEntering]; ok {
			pa.isMigrate = true
			for _, poolId := range poolIds {
				pa.poolTxs[common.Atoui(poolId)] = true
			}
		}
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
			"is_clp":        pa.isClp,
			"is_collect":    pa.isCollect,
			"is_migrate":    pa.isMigrate,
		})
		pa.poolInBlock[poolId] = true
	}
	pa.poolTxs = make(map[uint64]bool)
}

// AfterEndBlock checks for pools related ActiveProposal events and update the status accordingly. Update pool
// stats at the end.
func (pa *PoolAdapter) AfterEndBlock(
	ctx sdk.Context,
	_ abci.RequestEndBlock,
	evMap common.EvMap,
	kafka *[]common.Message,
) {
	if rawIds, ok := evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		// TODO: implement resolving multiple pool proposals
		if len(rawIds) > 1 {
			panic("Too many active prososals")
		}
		for idx, rawId := range rawIds {
			if evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalResult][idx] != govtypes.AttributeValueProposalPassed {
				continue
			}
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
			case *clptypes.CreateConcentratedLiquidityPoolsProposal:
				poolmanagerModuleAcc := pa.accountKeeper.GetModuleAccount(ctx, poolmanagertypes.ModuleName)
				poolCreatorAddress := poolmanagerModuleAcc.GetAddress()
				poolId := pa.poolmanagerKeeper.GetNextPoolId(ctx) - 1
				pool, _ := pa.clpKeeper.GetPool(ctx, poolId)
				concentratedPool, ok := pool.(clptypes.ConcentratedPoolExtension)
				if !ok {
					panic(fmt.Errorf("cannot decode concentrated pool from proposal, proposal => %d, pool id => %d", proposalId, poolId))
				}
				poolLiquidity, _ := pa.clpKeeper.GetTotalPoolLiquidity(ctx, poolId)
				common.AppendMessage(kafka, "NEW_OSMOSIS_POOL", common.JsDict{
					"id":                   poolId,
					"liquidity":            poolLiquidity,
					"type":                 concentratedPool.GetType(),
					"creator":              poolCreatorAddress,
					"is_superfluid":        false,
					"is_supported":         false,
					"swap_fee":             concentratedPool.GetSpreadFactor(ctx),
					"exit_fee":             "0",
					"future_pool_governor": "",
					"address":              concentratedPool.GetAddress().String(),
					"total_shares":         sdk.Coin{},
					"spread_factor":        concentratedPool.GetSpreadFactor(ctx),
					"tick_spacing":         concentratedPool.GetTickSpacing(),
				})
			case *clptypes.TickSpacingDecreaseProposal:
				for _, record := range content.PoolIdToTickSpacingRecords {
					common.AppendMessage(kafka, "UPDATE_POOL", common.JsDict{
						"id":           record.PoolId,
						"tick_spacing": record.NewTickSpacing,
					})
				}
			case *cosmwasmpooltypes.UploadCosmWasmPoolCodeAndWhiteListProposal:
				// TODO: refactor this logic when Osmosis unfork Cosmos SDK
				codeId := pa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeyLastCodeID) - 1
				codeInfo := pa.wasmKeeper.GetCodeInfo(ctx, codeId)
				if codeInfo == nil {
					break
				}
				addresses := make([]string, 0)
				switch codeInfo.InstantiateConfig.Permission {
				case wasmtypes.AccessTypeOnlyAddress:
					addresses = []string{codeInfo.InstantiateConfig.Address}
				case wasmtypes.AccessTypeAnyOfAddresses:
					addresses = codeInfo.InstantiateConfig.Addresses
				default:
					break
				}

				common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
					"id":                       codeId,
					"uploader":                 codeInfo.Creator,
					"contract_instantiated":    0,
					"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
					"access_config_addresses":  addresses,
				})
				common.AppendMessage(kafka, "NEW_CODE_PROPOSAL", common.JsDict{
					"code_id":         codeId,
					"proposal_id":     proposalId,
					"resolved_height": ctx.BlockHeight(),
				})
			case *cosmwasmpooltypes.MigratePoolContractsProposal:
				var newCodeId uint64
				if len(content.WASMByteCode) > 0 {
					newCodeId = pa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeyLastCodeID) - 1
				} else {
					newCodeId = content.NewCodeId
				}

				for _, poolId := range content.PoolIds {
					contractAddress, _, err := pa.cosmwasmpoolKeeper.GetCodeIdByPoolId(ctx, poolId)
					if err != nil {
						panic(err)
					}
					pa.updateContractVersion(ctx, contractAddress, kafka)
					common.AppendMessage(kafka, "UPDATE_CONTRACT_CODE_ID", common.JsDict{
						"contract": contractAddress.String(),
						"code_id":  newCodeId,
					})
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": contractAddress.String(),
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
					common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
						"contract_address": contractAddress.String(),
						"sender":           contractAddress.String(),
						"code_id":          newCodeId,
						"block_height":     ctx.BlockHeight(),
						"remark": common.JsDict{
							"type":      "governance",
							"operation": wasmtypes.ContractCodeHistoryOperationTypeMigrate.String(),
							"value":     proposalId,
						},
					})
				}
			}
		}
	}
	pa.flushUpdatePoolStats(ctx, kafka)
}

// getPoolIdFromDenom returns pool id if the provided denom is a pool share token. Otherwise, returns false.
func getPoolIdFromDenom(denom string) (string, bool) {
	if strings.HasPrefix(denom, "gamm/pool/") {
		return strings.Trim(denom, "gamm/pool/"), true
	}
	if strings.HasPrefix(denom, "cl/pool/") {
		return strings.Trim(denom, "cl/pool/"), true
	}

	return "", false
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

// updateContractVersion updates CW2 info of a contract using the query result.
func (pa *PoolAdapter) updateContractVersion(ctx sdk.Context, contractAddress sdk.AccAddress, kafka *[]common.Message) {
	contractInfo := pa.wasmKeeper.GetContractInfo(ctx, contractAddress)
	rawContractVersion := pa.wasmKeeper.QueryRaw(ctx, contractAddress, []byte("contract_info"))
	var contractVersion ContractVersion
	err := json.Unmarshal(rawContractVersion, &contractVersion)
	if err != nil {
		return
	}
	common.AppendMessage(kafka, "UPDATE_CW2_INFO", common.JsDict{
		"code_id":      contractInfo.CodeID,
		"cw2_contract": contractVersion.Contract,
		"cw2_version":  contractVersion.Version,
	})
}

// handleCreatePoolEvents handles PoolCreated events and processes new pools in the current transaction.
func (pa *PoolAdapter) handleCreatePoolEvents(ctx sdk.Context, txHash []byte, sender sdk.AccAddress, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[poolmanagertypes.TypeEvtPoolCreated+"."+gammtypes.AttributeKeyPoolId]; ok {
		newPool := make(map[uint64]bool)
		for _, rawId := range rawIds {
			newPool[common.Atoui(rawId)] = true
		}
		for poolId := range newPool {
			poolInfo, _ := pa.poolmanagerKeeper.GetPool(ctx, poolId)
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
					"swap_fee":             pool.GetSpreadFactor(ctx).String(),
					"exit_fee":             pool.GetExitFee(ctx).String(),
					"future_pool_governor": pool.FuturePoolGovernor,
					"weight":               weights,
					"address":              pool.GetAddress().String(),
					"total_shares":         pool.TotalShares,
				}
				if smoothWeightChangeParams := pool.PoolParams.GetSmoothWeightChangeParams(); smoothWeightChangeParams != nil {
					poolMsg["smooth_weight_change_params"] = smoothWeightChangeParams
				}
				common.AppendMessage(kafka, "NEW_OSMOSIS_POOL", poolMsg)
			case *stableswap.Pool:
				common.AppendMessage(kafka, "NEW_OSMOSIS_POOL", common.JsDict{
					"id":                        poolId,
					"liquidity":                 pool.GetTotalPoolLiquidity(ctx),
					"type":                      pool.GetType(),
					"creator":                   sender.String(),
					"create_tx":                 txHash,
					"is_superfluid":             false,
					"is_supported":              false,
					"swap_fee":                  pool.GetSpreadFactor(ctx).String(),
					"exit_fee":                  pool.GetExitFee(ctx).String(),
					"future_pool_governor":      pool.FuturePoolGovernor,
					"scaling_factors":           pool.GetScalingFactors(),
					"scaling_factor_controller": pool.ScalingFactorController,
					"address":                   pool.GetAddress().String(),
					"total_shares":              pool.TotalShares,
				})
			case *concentratedpool.Pool:
				token0 := pool.GetToken0()
				token1 := pool.GetToken1()
				poolLiquidity := []sdk.Coin{sdk.NewCoin(token0, sdk.ZeroInt()), sdk.NewCoin(token1, sdk.ZeroInt())}
				common.AppendMessage(kafka, "NEW_OSMOSIS_POOL", common.JsDict{
					"id":                   poolId,
					"liquidity":            poolLiquidity,
					"type":                 pool.GetType(),
					"creator":              sender.String(),
					"create_tx":            txHash,
					"is_superfluid":        false,
					"is_supported":         false,
					"swap_fee":             pool.GetSpreadFactor(ctx),
					"exit_fee":             "0",
					"future_pool_governor": "",
					"address":              pool.GetAddress().String(),
					"total_shares":         sdk.Coin{},
					"spread_factor":        pool.GetSpreadFactor(ctx),
					"tick_spacing":         pool.GetTickSpacing(),
				})
			case *cosmwasmpool.Pool:
				pa.cosmwasmContractAddrs[pool.GetContractAddress()] = poolId
				common.AppendMessage(kafka, "NEW_OSMOSIS_POOL", common.JsDict{
					"id":                   poolId,
					"liquidity":            pool.GetTotalPoolLiquidity(ctx),
					"type":                 pool.GetType(),
					"creator":              sender,
					"create_tx":            txHash,
					"is_superfluid":        false,
					"is_supported":         false,
					"swap_fee":             pool.GetSpreadFactor(ctx),
					"exit_fee":             "0",
					"future_pool_governor": "",
					"address":              pool.GetAddress(),
					"total_shares":         sdk.Coin{},
					"contract_address":     pool.GetContractAddress(),
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
	if positionIds, ok := evMap[superfluidtypes.TypeEvtAddToConcentratedLiquiditySuperfluidPosition+"."+superfluidtypes.AttributePositionId]; ok {
		for _, positionId := range positionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, common.Atoui(positionId)); err == nil {
				pa.isSuperfluidTx = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	}
	if poolIds, ok := evMap[superfluidtypes.TypeEvtUnlockAndMigrateShares+"."+superfluidtypes.AttributeKeyPoolIdEntering]; ok {
		pa.isMigrate = true
		for _, poolId := range poolIds {
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if poolIds, ok := evMap[superfluidtypes.TypeEvtUnlockAndMigrateShares+"."+superfluidtypes.AttributeKeyPoolIdLeaving]; ok {
		pa.isMigrate = true
		for _, poolId := range poolIds {
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if positionIds, ok := evMap[superfluidtypes.TypeEvtCreateFullRangePositionAndSFDelegate+"."+superfluidtypes.AttributePositionId]; ok {
		for _, positionId := range positionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, common.Atoui(positionId)); err == nil {
				pa.isSuperfluidTx = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	}
	if poolIds, ok := evMap[clptypes.TypeEvtCreatePosition+"."+clptypes.AttributeKeyPoolId]; ok {
		for _, poolId := range poolIds {
			pa.isClp = true
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if poolIds, ok := evMap[clptypes.TypeEvtWithdrawPosition+"."+clptypes.AttributeKeyPoolId]; ok {
		for _, poolId := range poolIds {
			pa.isClp = true
			pa.poolTxs[common.Atoui(poolId)] = true
		}
	}
	if positionIds, ok := evMap[clptypes.TypeEvtAddToPosition+"."+clptypes.AttributeKeyPositionId]; ok {
		for _, positionId := range positionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, common.Atoui(positionId)); err == nil {
				pa.isClp = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	}
	if positionIds, ok := evMap[clptypes.TypeEvtCollectSpreadRewards+"."+clptypes.AttributeKeyPositionId]; ok {
		for _, positionId := range positionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, common.Atoui(positionId)); err == nil {
				pa.isCollect = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	}
	if positionIds, ok := evMap[clptypes.TypeEvtCollectIncentives+"."+clptypes.AttributeKeyPositionId]; ok {
		for _, positionId := range positionIds {
			if position, err := pa.clpKeeper.GetPosition(ctx, common.Atoui(positionId)); err == nil {
				pa.isCollect = true
				pa.poolTxs[position.PoolId] = true
			}
		}
	}
	if contractAddresses, ok := evMap[wasmtypes.EventTypeExecute+"."+wasmtypes.AttributeKeyContractAddr]; ok {
		for _, contract := range contractAddresses {
			if poolId, found := pa.cosmwasmContractAddrs[contract]; found {
				pa.poolTxs[poolId] = true
			}
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
		poolInfo, _ := pa.poolmanagerKeeper.GetPool(ctx, poolId)
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
		case *concentratedpool.Pool:
			poolLiquidity, _ := pa.clpKeeper.GetTotalPoolLiquidity(ctx, poolId)
			common.AppendMessage(kafka, "UPDATE_POOL", common.JsDict{
				"id":        poolId,
				"liquidity": poolLiquidity,
			})
		case *cosmwasmpool.Pool:
			common.AppendMessage(kafka, "UPDATE_POOL", common.JsDict{
				"id":        poolId,
				"liquidity": pool.GetTotalPoolLiquidity(ctx),
			})
		default:
			panic("cannot handle pool type")
		}
	}
	pa.poolInBlock = make(map[uint64]bool)
}
