package emitter

import (
	"encoding/json"
	"fmt"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/osmosis-labs/osmosis/v21/app/params"
	"github.com/osmosis-labs/osmosis/v21/hooks/common"
)

var (
	_ Adapter = &WasmAdapter{}
)

// WasmAdapter defines a struct containing required keepers and flags to process the CosmWasm related hook.
// It implements Adapter interface.
type WasmAdapter struct {
	// tx flags
	contractTxs     map[string]bool
	isStoreCodeTx   bool
	isInstantiateTx bool
	isExecuteTx     bool
	isSendTx        bool
	isUpdateAdmin   bool
	isClearAdmin    bool
	isMigrate       bool
	isIBC           bool

	// wasm params
	lastInstanceID                uint64
	maxCodeIDfromTx               uint64
	maxInstanceIDfromTx           uint64
	codeIDInstantiateFromProposal []uint64

	// keepers
	wasmKeeper *wasmkeeper.Keeper
	govKeeper  *govkeeper.Keeper
}

// NewWasmAdapter creates new WasmAdapter instance that will be added to the emitter hook adapters.
func NewWasmAdapter(wasmKeeper *wasmkeeper.Keeper, govKeeper *govkeeper.Keeper) *WasmAdapter {
	return &WasmAdapter{
		contractTxs:     make(map[string]bool),
		isStoreCodeTx:   false,
		isInstantiateTx: false,
		isExecuteTx:     false,
		isSendTx:        false,
		isUpdateAdmin:   false,
		isClearAdmin:    false,
		isMigrate:       false,
		isIBC:           false,
		lastInstanceID:  0,
		wasmKeeper:      wasmKeeper,
		govKeeper:       govKeeper,
	}
}

// AfterInitChain extracts codes and contracts from the given genesis state.
func (wa *WasmAdapter) AfterInitChain(ctx sdk.Context, encodingConfig params.EncodingConfig, genesisState map[string]json.RawMessage, kafka *[]common.Message) {
	var wasmGenesis wasmtypes.GenesisState
	if genesisState[wasmtypes.ModuleName] != nil {
		encodingConfig.Marshaler.MustUnmarshalJSON(genesisState[wasmtypes.ModuleName], &wasmGenesis)
	}
	for _, code := range wasmGenesis.Codes {
		addresses := make([]string, 0)
		switch code.CodeInfo.InstantiateConfig.Permission {
		case wasmtypes.AccessTypeAnyOfAddresses:
			addresses = code.CodeInfo.InstantiateConfig.Addresses
		}
		uploader, _ := sdk.AccAddressFromBech32(code.CodeInfo.Creator)
		common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
			"id":                       code.CodeID,
			"uploader":                 uploader,
			"contract_instantiated":    0,
			"access_config_permission": code.CodeInfo.InstantiateConfig.Permission.String(),
			"access_config_addresses":  addresses,
			"hash":                     code.CodeInfo.CodeHash,
		})
	}

	wa.wasmKeeper.IterateContractInfo(ctx, func(addr sdk.AccAddress, contractInfo types.ContractInfo) bool {
		histories := wa.wasmKeeper.GetContractHistory(ctx, addr)
		wa.updateContractVersion(ctx, addr.String(), kafka)
		for _, history := range histories {
			if history.Operation == types.ContractCodeHistoryOperationTypeInit {
				common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
					"address":           addr,
					"code_id":           history.CodeID,
					"init_msg":          string(history.Msg),
					"init_by":           contractInfo.Creator,
					"contract_executed": 0,
					"label":             contractInfo.Label,
					"admin":             contractInfo.Admin,
				})
				common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
					"id": contractInfo.CodeID,
				})
				common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
					"contract_address": addr,
					"sender":           contractInfo.Creator,
					"code_id":          history.CodeID,
					"block_height":     ctx.BlockHeight(),
					"remark": common.JsDict{
						"type":      "genesis",
						"operation": history.Operation.String(),
					},
				})
			}
		}
		return false
	})
}

// AfterBeginBlock assigns updated values to the adapter variables before processing transactions in each block.
func (wa *WasmAdapter) AfterBeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
	wa.maxCodeIDfromTx = wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
	wa.maxInstanceIDfromTx = wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceInstanceID)
	wa.codeIDInstantiateFromProposal = make([]uint64, 0)
}

// PreDeliverTx sets the necessary maps and flags to the default value before processing each transaction.
func (wa *WasmAdapter) PreDeliverTx() {
	wa.isStoreCodeTx = false
	wa.isInstantiateTx = false
	wa.isExecuteTx = false
	wa.isClearAdmin = false
	wa.isUpdateAdmin = false
	wa.isMigrate = false
	wa.contractTxs = make(map[string]bool)
}

// CheckMsg checks the message type and extracts message values to WasmAdapter maps and flags accordingly.
func (wa *WasmAdapter) CheckMsg(_ sdk.Context, msg sdk.Msg) {
	switch msg := msg.(type) {
	case *wasmtypes.MsgStoreCode:
		wa.isStoreCodeTx = true
	case *wasmtypes.MsgInstantiateContract:
		wa.isInstantiateTx = true
	case *wasmtypes.MsgExecuteContract:
		wa.isExecuteTx = true
		wa.contractTxs[msg.Contract] = false
	case *wasmtypes.MsgUpdateAdmin:
		wa.isUpdateAdmin = true
		wa.contractTxs[msg.Contract] = false
	case *wasmtypes.MsgClearAdmin:
		wa.isClearAdmin = true
		wa.contractTxs[msg.Contract] = false
	case *wasmtypes.MsgMigrateContract:
		wa.isMigrate = true
		wa.contractTxs[msg.Contract] = false
	case *wasmtypes.MsgInstantiateContract2:
		wa.isInstantiateTx = true
	case *wasmtypes.MsgUpdateContractLabel:
		wa.contractTxs[msg.Contract] = false
	case *wasmtypes.MsgStoreAndInstantiateContract:
		wa.isStoreCodeTx = true
		wa.isInstantiateTx = true
	case *wasmtypes.MsgStoreAndMigrateContract:
		wa.isStoreCodeTx = true
		wa.isMigrate = true
	case *channeltypes.MsgRecvPacket:
		if contractAddr, err := wasm.ContractFromPortID(msg.Packet.DestinationPort); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	case *channeltypes.MsgChannelOpenAck:
		if contractAddr, err := wasm.ContractFromPortID(msg.PortId); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	case *channeltypes.MsgChannelOpenTry:
		if contractAddr, err := wasm.ContractFromPortID(msg.PortId); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	case *channeltypes.MsgChannelOpenConfirm:
		if contractAddr, err := wasm.ContractFromPortID(msg.PortId); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	case *channeltypes.MsgTimeout:
		if contractAddr, err := wasm.ContractFromPortID(msg.Packet.SourcePort); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	case *channeltypes.MsgAcknowledgement:
		if contractAddr, err := wasm.ContractFromPortID(msg.Packet.SourcePort); err == nil {
			wa.contractTxs[contractAddr.String()] = true
		}
	}
}

// HandleMsgEvents checks for new code, new contract, contract execution, contract migration and new contract related
// proposal events emitted from the given message.
func (wa *WasmAdapter) HandleMsgEvents(ctx sdk.Context, txHash []byte, msg sdk.Msg, evMap common.EvMap, detail common.JsDict, kafka *[]common.Message) {
	wa.updateNewCodeEvents(ctx, txHash, msg, evMap, kafka)
	wa.updateNewContractEvents(ctx, txHash, evMap, kafka)
	wa.updateContractExecuteEvents(evMap)
	wa.updateContractSudoEvents(evMap)
	wa.updateContractProposalEvents(ctx, evMap, kafka)
	wa.updateMigrateContractEvents(ctx, txHash, msg, evMap, kafka)
	wa.updateContractAdminEvents(evMap, kafka)
	wa.updateContractLabelEvents(evMap, kafka)
	wa.updateCodeInstantiateConfigEvents(ctx, evMap, kafka)

	switch msg := msg.(type) {
	case *types.MsgStoreCode:
		detail["id"] = common.Atoui(evMap[types.EventTypeStoreCode+"."+types.AttributeKeyCodeID][0])
	case *types.MsgInstantiateContract:
		contracts := evMap[types.EventTypeInstantiate+"."+types.AttributeKeyContractAddr]
		detail["_contract_address"] = contracts[0]
		detail["_contract_addresses"] = contracts
	case *types.MsgInstantiateContract2:
		contracts := evMap[types.EventTypeInstantiate+"."+types.AttributeKeyContractAddr]
		detail["_contract_address"] = contracts[0]
		detail["_contract_addresses"] = contracts
	case *types.MsgExecuteContract:
		detail["msg_json"] = string(msg.Msg)
	case *types.MsgClearAdmin:
		common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
			"contract": msg.Contract,
			"admin":    "",
		})
	case *types.MsgUpdateAdmin:
		common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
			"contract": msg.Contract,
			"admin":    msg.NewAdmin,
		})
	}
}

// PostDeliverTx appends contract transactions into the Kafka message array and assigns WasmAdapter flags to the
// interface to be written to the message queue.
func (wa *WasmAdapter) PostDeliverTx(ctx sdk.Context, txHash []byte, txDict common.JsDict, kafka *[]common.Message) {
	for contractAddr := range wa.contractTxs {
		common.AppendMessage(kafka, "NEW_CONTRACT_TRANSACTION", common.JsDict{
			"contract_address": contractAddr,
			"tx_hash":          txHash,
			"is_instantiate":   wa.isInstantiateTx,
		})
	}
	wa.contractTxs = make(map[string]bool)

	txDict["is_store_code"] = wa.isStoreCodeTx
	txDict["is_instantiate"] = wa.isInstantiateTx
	txDict["is_execute"] = wa.isExecuteTx
	txDict["is_update_admin"] = wa.isUpdateAdmin
	txDict["is_clear_admin"] = wa.isClearAdmin
	txDict["is_migrate"] = wa.isMigrate
	wa.lastInstanceID = wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceInstanceID)
}

// AfterEndBlock checks for wasm related ActiveProposal events and process them accordingly.
func (wa *WasmAdapter) AfterEndBlock(ctx sdk.Context, _ abci.RequestEndBlock, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for idx, rawId := range rawIds {
			if evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalResult][idx] != govtypes.AttributeValueProposalPassed {
				continue
			}
			proposalId := common.Atoui(rawId)
			proposal, _ := wa.govKeeper.GetProposal(ctx, proposalId)
			msgs, _ := proposal.GetMsgs()
			for _, msg := range msgs {
				switch msg := msg.(type) {
				case *govv1types.MsgExecLegacyContent:
					content, err := govv1types.LegacyContentFromMessage(msg)
					if err != nil {
						panic(err)
					}
					switch content := content.(type) {
					case *wasmtypes.StoreCodeProposal:
						lastCodeID := wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
						for id := wa.maxCodeIDfromTx; id < lastCodeID; id++ {
							codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, id)
							if codeInfo == nil {
								break
							}
							addresses := make([]string, 0)
							switch codeInfo.InstantiateConfig.Permission {
							case wasmtypes.AccessTypeAnyOfAddresses:
								addresses = codeInfo.InstantiateConfig.Addresses
							default:
								break
							}

							common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
								"id":                       id,
								"uploader":                 codeInfo.Creator,
								"contract_instantiated":    0,
								"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
								"access_config_addresses":  addresses,
								"hash":                     codeInfo.CodeHash,
							})
							common.AppendMessage(kafka, "NEW_CODE_PROPOSAL", common.JsDict{
								"code_id":         id,
								"proposal_id":     proposalId,
								"resolved_height": ctx.BlockHeight(),
							})
						}
					case *wasmtypes.InstantiateContractProposal:
						contractAddr := wasmkeeper.BuildContractAddressClassic(content.CodeID, wa.lastInstanceID)
						wa.updateContractVersion(ctx, contractAddr.String(), kafka)
						histories := wa.wasmKeeper.GetContractHistory(ctx, contractAddr)
						info := wa.wasmKeeper.GetContractInfo(ctx, contractAddr)
						for _, history := range histories {
							if history.Operation == wasmtypes.ContractCodeHistoryOperationTypeInit {
								common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
									"address":           contractAddr,
									"code_id":           history.CodeID,
									"init_msg":          string(history.Msg),
									"init_by":           info.Creator,
									"contract_executed": 0,
									"label":             info.Label,
									"admin":             info.Admin,
								})
								common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
									"id": info.CodeID,
								})
								common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
									"contract_address": contractAddr,
									"proposal_id":      proposalId,
									"resolved_height":  ctx.BlockHeight(),
								})
								common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
									"contract_address": contractAddr,
									"sender":           info.Creator,
									"code_id":          history.CodeID,
									"block_height":     ctx.BlockHeight(),
									"remark": common.JsDict{
										"type":      "governance",
										"operation": wasmtypes.ContractCodeHistoryOperationTypeInit.String(),
										"value":     proposalId,
									},
								})
							}
							break
						}
					case *wasmtypes.MigrateContractProposal:
						wa.updateContractVersion(ctx, content.Contract, kafka)
						common.AppendMessage(kafka, "UPDATE_CONTRACT_CODE_ID", common.JsDict{
							"contract": content.Contract,
							"code_id":  content.CodeID,
						})
						common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
							"resolved_height":  ctx.BlockHeight(),
						})
						common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
							"contract_address": content.Contract,
							"sender":           content.Contract,
							"code_id":          content.CodeID,
							"block_height":     ctx.BlockHeight(),
							"remark": common.JsDict{
								"type":      "governance",
								"operation": wasmtypes.ContractCodeHistoryOperationTypeMigrate.String(),
								"value":     proposalId,
							},
						})
					case *wasmtypes.UpdateAdminProposal:
						common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
							"contract": content.Contract,
							"admin":    content.NewAdmin,
						})
						common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
							"resolved_height":  ctx.BlockHeight(),
						})
					case *wasmtypes.ClearAdminProposal:
						common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
							"contract": content.Contract,
							"admin":    "",
						})
						common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
							"resolved_height":  ctx.BlockHeight(),
						})
					case *wasmtypes.ExecuteContractProposal:
						common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
							"resolved_height":  ctx.BlockHeight(),
						})
					}
				case *wasmtypes.MsgStoreCode:
					lastCodeID := wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
					for id := wa.maxCodeIDfromTx; id < lastCodeID; id++ {
						codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, id)
						if codeInfo == nil {
							break
						}
						addresses := make([]string, 0)
						switch codeInfo.InstantiateConfig.Permission {
						case wasmtypes.AccessTypeAnyOfAddresses:
							addresses = codeInfo.InstantiateConfig.Addresses
						default:
							break
						}

						common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
							"id":                       id,
							"uploader":                 codeInfo.Creator,
							"contract_instantiated":    0,
							"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
							"access_config_addresses":  addresses,
							"hash":                     codeInfo.CodeHash,
						})
						common.AppendMessage(kafka, "NEW_CODE_PROPOSAL", common.JsDict{
							"code_id":         id,
							"proposal_id":     proposalId,
							"resolved_height": ctx.BlockHeight(),
						})
					}
				case *wasmtypes.MsgInstantiateContract:
					contractAddr := wasmkeeper.BuildContractAddressClassic(msg.CodeID, wa.lastInstanceID)
					wa.updateContractVersion(ctx, contractAddr.String(), kafka)
					histories := wa.wasmKeeper.GetContractHistory(ctx, contractAddr)
					info := wa.wasmKeeper.GetContractInfo(ctx, contractAddr)
					for _, history := range histories {
						if history.Operation == wasmtypes.ContractCodeHistoryOperationTypeInit {
							common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
								"address":           contractAddr,
								"code_id":           history.CodeID,
								"init_msg":          string(history.Msg),
								"init_by":           info.Creator,
								"contract_executed": 0,
								"label":             info.Label,
								"admin":             info.Admin,
							})
							common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
								"id": info.CodeID,
							})
							common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
								"contract_address": contractAddr,
								"proposal_id":      proposalId,
								"resolved_height":  ctx.BlockHeight(),
							})
							common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
								"contract_address": contractAddr,
								"sender":           info.Creator,
								"code_id":          history.CodeID,
								"block_height":     ctx.BlockHeight(),
								"remark": common.JsDict{
									"type":      "governance",
									"operation": wasmtypes.ContractCodeHistoryOperationTypeInit.String(),
									"value":     proposalId,
								},
							})
						}
						break
					}
				case *wasmtypes.MsgInstantiateContract2:
					contractAddr := wasmkeeper.BuildContractAddressClassic(msg.CodeID, wa.lastInstanceID)
					wa.updateContractVersion(ctx, contractAddr.String(), kafka)
					histories := wa.wasmKeeper.GetContractHistory(ctx, contractAddr)
					info := wa.wasmKeeper.GetContractInfo(ctx, contractAddr)
					for _, history := range histories {
						if history.Operation == wasmtypes.ContractCodeHistoryOperationTypeInit {
							common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
								"address":           contractAddr,
								"code_id":           history.CodeID,
								"init_msg":          string(history.Msg),
								"init_by":           info.Creator,
								"contract_executed": 0,
								"label":             info.Label,
								"admin":             info.Admin,
							})
							common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
								"id": info.CodeID,
							})
							common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
								"contract_address": contractAddr,
								"proposal_id":      proposalId,
								"resolved_height":  ctx.BlockHeight(),
							})
							common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
								"contract_address": contractAddr,
								"sender":           info.Creator,
								"code_id":          history.CodeID,
								"block_height":     ctx.BlockHeight(),
								"remark": common.JsDict{
									"type":      "governance",
									"operation": wasmtypes.ContractCodeHistoryOperationTypeInit.String(),
									"value":     proposalId,
								},
							})
						}
						break
					}
				case *wasmtypes.MsgExecuteContract:
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
				case *wasmtypes.MsgMigrateContract:
					wa.updateContractVersion(ctx, msg.Contract, kafka)
					common.AppendMessage(kafka, "UPDATE_CONTRACT_CODE_ID", common.JsDict{
						"contract": msg.Contract,
						"code_id":  msg.CodeID,
					})
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
					common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
						"contract_address": msg.Contract,
						"sender":           msg.Contract,
						"code_id":          msg.CodeID,
						"block_height":     ctx.BlockHeight(),
						"remark": common.JsDict{
							"type":      "governance",
							"operation": wasmtypes.ContractCodeHistoryOperationTypeMigrate.String(),
							"value":     proposalId,
						},
					})
				case *wasmtypes.MsgUpdateAdmin:
					common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
						"contract": msg.Contract,
						"admin":    msg.NewAdmin,
					})
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
				case *wasmtypes.MsgClearAdmin:
					common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
						"contract": msg.Contract,
						"admin":    "",
					})
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
				case *wasmtypes.MsgSudoContract:
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
				case *wasmtypes.MsgStoreAndInstantiateContract:
					lastCodeID := wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
					for id := wa.maxCodeIDfromTx; id < lastCodeID; id++ {
						codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, id)
						if codeInfo == nil {
							break
						}
						addresses := make([]string, 0)
						switch codeInfo.InstantiateConfig.Permission {
						case wasmtypes.AccessTypeAnyOfAddresses:
							addresses = codeInfo.InstantiateConfig.Addresses
						default:
							break
						}

						common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
							"id":                       id,
							"uploader":                 codeInfo.Creator,
							"contract_instantiated":    0,
							"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
							"access_config_addresses":  addresses,
							"hash":                     codeInfo.CodeHash,
						})
						common.AppendMessage(kafka, "NEW_CODE_PROPOSAL", common.JsDict{
							"code_id":         id,
							"proposal_id":     proposalId,
							"resolved_height": ctx.BlockHeight(),
						})

						contractAddr := wasmkeeper.BuildContractAddressClassic(id, wa.lastInstanceID)
						wa.updateContractVersion(ctx, contractAddr.String(), kafka)
						histories := wa.wasmKeeper.GetContractHistory(ctx, contractAddr)
						info := wa.wasmKeeper.GetContractInfo(ctx, contractAddr)
						for _, history := range histories {
							if history.Operation == wasmtypes.ContractCodeHistoryOperationTypeInit {
								common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
									"address":           contractAddr,
									"code_id":           history.CodeID,
									"init_msg":          string(history.Msg),
									"init_by":           info.Creator,
									"contract_executed": 0,
									"label":             info.Label,
									"admin":             info.Admin,
								})
								common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
									"id": info.CodeID,
								})
								common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
									"contract_address": contractAddr,
									"proposal_id":      proposalId,
									"resolved_height":  ctx.BlockHeight(),
								})
								common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
									"contract_address": contractAddr,
									"sender":           info.Creator,
									"code_id":          history.CodeID,
									"block_height":     ctx.BlockHeight(),
									"remark": common.JsDict{
										"type":      "governance",
										"operation": wasmtypes.ContractCodeHistoryOperationTypeInit.String(),
										"value":     proposalId,
									},
								})
							}
							break
						}
					}
				case *wasmtypes.MsgStoreAndMigrateContract:
					lastCodeID := wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
					for id := wa.maxCodeIDfromTx; id < lastCodeID; id++ {
						codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, id)
						if codeInfo == nil {
							break
						}
						addresses := make([]string, 0)
						switch codeInfo.InstantiateConfig.Permission {
						case wasmtypes.AccessTypeAnyOfAddresses:
							addresses = codeInfo.InstantiateConfig.Addresses
						default:
							break
						}

						common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
							"id":                       id,
							"uploader":                 codeInfo.Creator,
							"contract_instantiated":    0,
							"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
							"access_config_addresses":  addresses,
							"hash":                     codeInfo.CodeHash,
						})
						common.AppendMessage(kafka, "NEW_CODE_PROPOSAL", common.JsDict{
							"code_id":         id,
							"proposal_id":     proposalId,
							"resolved_height": ctx.BlockHeight(),
						})
						wa.updateContractVersion(ctx, msg.Contract, kafka)
						common.AppendMessage(kafka, "UPDATE_CONTRACT_CODE_ID", common.JsDict{
							"contract": msg.Contract,
							"code_id":  id,
						})
						common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": msg.Contract,
							"proposal_id":      proposalId,
							"resolved_height":  ctx.BlockHeight(),
						})
						common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
							"contract_address": msg.Contract,
							"sender":           msg.Contract,
							"code_id":          id,
							"block_height":     ctx.BlockHeight(),
							"remark": common.JsDict{
								"type":      "governance",
								"operation": wasmtypes.ContractCodeHistoryOperationTypeMigrate.String(),
								"value":     proposalId,
							},
						})
					}
				case *wasmtypes.MsgUpdateContractLabel:
					common.AppendMessage(kafka, "UPDATE_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
						"resolved_height":  ctx.BlockHeight(),
					})
					common.AppendMessage(kafka, "UPDATE_CONTRACT_LABEL", common.JsDict{
						"contract": msg.Contract,
						"label":    msg.NewLabel,
					})
				}
			}
		}
	}
}

// updateContractProposalEvents handles wasm related SubmitProposal events that might be emitted from the transaction.
func (wa *WasmAdapter) updateContractProposalEvents(ctx sdk.Context, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[govtypes.EventTypeSubmitProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			proposalId := common.Atoui(rawId)
			proposal, _ := wa.govKeeper.GetProposal(ctx, proposalId)
			msgs, _ := proposal.GetMsgs()
			for _, msg := range msgs {
				switch msg := msg.(type) {
				case *govv1types.MsgExecLegacyContent:
					content, err := govv1types.LegacyContentFromMessage(msg)
					if err != nil {
						panic(err)
					}
					switch content := content.(type) {
					case *wasmtypes.MigrateContractProposal:
						common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
						})
					case *wasmtypes.SudoContractProposal:
						common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
						})
					case *wasmtypes.ExecuteContractProposal:
						common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
						})
					case *wasmtypes.UpdateAdminProposal:
						common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
						})
					case *wasmtypes.ClearAdminProposal:
						common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
							"contract_address": content.Contract,
							"proposal_id":      proposalId,
						})
					}
				case *wasmtypes.MsgMigrateContract:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})

				case *wasmtypes.MsgExecuteContract:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				case *wasmtypes.MsgUpdateAdmin:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				case *wasmtypes.MsgClearAdmin:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				case *wasmtypes.MsgSudoContract:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				case *wasmtypes.MsgStoreAndMigrateContract:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				case *wasmtypes.MsgUpdateContractLabel:
					common.AppendMessage(kafka, "NEW_CONTRACT_PROPOSAL", common.JsDict{
						"contract_address": msg.Contract,
						"proposal_id":      proposalId,
					})
				}
			}
		}
	}
}

// updateNewCodeEvents handles StoreCode events that might be emitted from the transaction.
func (wa *WasmAdapter) updateNewCodeEvents(ctx sdk.Context, txHash []byte, msg sdk.Msg, evMap common.EvMap, kafka *[]common.Message) {
	if rawIDs, ok := evMap[types.EventTypeStoreCode+"."+types.AttributeKeyCodeID]; ok {
		for _, rawId := range rawIDs {
			id := common.Atoui(rawId)
			codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, id)
			wa.maxCodeIDfromTx = wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceCodeID)
			addresses := make([]string, 0)
			switch codeInfo.InstantiateConfig.Permission {
			case wasmtypes.AccessTypeAnyOfAddresses:
				addresses = codeInfo.InstantiateConfig.Addresses
			default:
				break
			}
			common.AppendMessage(kafka, "NEW_CODE", common.JsDict{
				"id":                       id,
				"uploader":                 msg.GetSigners()[0],
				"contract_instantiated":    0,
				"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
				"access_config_addresses":  addresses,
				"tx_hash":                  txHash,
				"hash":                     codeInfo.CodeHash,
			})
		}
	}
}

// ContractVersion is a type for storing CW2 info of a contract.
type ContractVersion struct {
	Contract string
	Version  string
}

// updateContractVersion updates CW2 info of a contract using the query result.
func (wa *WasmAdapter) updateContractVersion(ctx sdk.Context, contract string, kafka *[]common.Message) {
	contractAddress, _ := sdk.AccAddressFromBech32(contract)
	contractInfo := wa.wasmKeeper.GetContractInfo(ctx, contractAddress)
	rawContractVersion := wa.wasmKeeper.QueryRaw(ctx, contractAddress, []byte("contract_info"))
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

// updateNewContractEvents handles contract Instantiate events that might be emitted from the transaction.
func (wa *WasmAdapter) updateNewContractEvents(ctx sdk.Context, txHash []byte, evMap common.EvMap, kafka *[]common.Message) {
	if events, ok := evMap[types.EventTypeInstantiate+"."+types.AttributeKeyContractAddr]; ok {
		wa.maxInstanceIDfromTx = wa.wasmKeeper.PeekAutoIncrementID(ctx, wasmtypes.KeySequenceInstanceID)
		for _, contractAddr := range events {
			addr, _ := sdk.AccAddressFromBech32(contractAddr)
			histories := wa.wasmKeeper.GetContractHistory(ctx, addr)
			info := wa.wasmKeeper.GetContractInfo(ctx, addr)
			wa.updateContractVersion(ctx, contractAddr, kafka)
			for _, history := range histories {
				if history.Operation == types.ContractCodeHistoryOperationTypeInit {
					common.AppendMessage(kafka, "NEW_CONTRACT", common.JsDict{
						"address":           contractAddr,
						"code_id":           history.CodeID,
						"init_msg":          string(history.Msg),
						"tx_hash":           txHash,
						"init_by":           info.Creator,
						"contract_executed": 0,
						"label":             info.Label,
						"admin":             info.Admin,
					})
					common.AppendMessage(kafka, "UPDATE_CODE", common.JsDict{
						"id": info.CodeID,
					})
					common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
						"contract_address": contractAddr,
						"sender":           info.Creator,
						"code_id":          history.CodeID,
						"block_height":     ctx.BlockHeight(),
						"remark": common.JsDict{
							"type":      "transaction",
							"operation": history.Operation.String(),
							"value":     fmt.Sprintf("%X", txHash),
						},
					})
					wa.contractTxs[contractAddr] = false
					break
				}
			}
		}
	}
}

// updateMigrateContractEvents handles contract Migrate events that might be emitted from the transaction.
func (wa *WasmAdapter) updateMigrateContractEvents(ctx sdk.Context, txHash []byte, msg sdk.Msg, evMap common.EvMap, kafka *[]common.Message) {
	if events, ok := evMap[types.EventTypeMigrate+"."+types.AttributeKeyContractAddr]; ok {
		for idx, contractAddr := range events {
			wa.updateContractVersion(ctx, contractAddr, kafka)
			codeID := common.Atoui(evMap[types.EventTypeMigrate+"."+types.AttributeKeyCodeID][idx])
			common.AppendMessage(kafka, "UPDATE_CONTRACT_CODE_ID", common.JsDict{
				"contract": contractAddr,
				"code_id":  codeID,
			})
			common.AppendMessage(kafka, "NEW_CONTRACT_HISTORY", common.JsDict{
				"contract_address": contractAddr,
				"sender":           msg.GetSigners()[0],
				"code_id":          codeID,
				"block_height":     ctx.BlockHeight(),
				"remark": common.JsDict{
					"type":      "transaction",
					"operation": wasmtypes.ContractCodeHistoryOperationTypeMigrate.String(),
					"value":     fmt.Sprintf("%X", txHash),
				},
			})
			wa.contractTxs[contractAddr] = false
		}
	}
}

// updateContractExecuteEvents handles contract Execute events that might be emitted from the transaction.
func (wa *WasmAdapter) updateContractExecuteEvents(evMap common.EvMap) {
	for _, contract := range evMap[types.EventTypeExecute+"."+types.AttributeKeyContractAddr] {
		wa.contractTxs[contract] = false
	}
}

// updateContractSudoEvents handles contract Sudo events that might be emitted from the transaction.
func (wa *WasmAdapter) updateContractSudoEvents(evMap common.EvMap) {
	for _, contract := range evMap[types.EventTypeSudo+"."+types.AttributeKeyContractAddr] {
		wa.contractTxs[contract] = false
	}
}

// updateContractAdminEvents handles contract admin update events that might be emitted from the transaction.
func (wa *WasmAdapter) updateContractAdminEvents(evMap common.EvMap, kafka *[]common.Message) {
	for idx, contract := range evMap[wasmtypes.EventTypeUpdateContractAdmin+"."+wasmtypes.AttributeKeyContractAddr] {
		common.AppendMessage(kafka, "UPDATE_CONTRACT_ADMIN", common.JsDict{
			"contract": contract,
			"admin":    evMap[wasmtypes.EventTypeUpdateContractAdmin+"."+wasmtypes.AttributeKeyNewAdmin][idx],
		})
		wa.contractTxs[contract] = false
	}
}

// updateContractLabelEvents handles contract label update events that might be emitted from the transaction.
func (wa *WasmAdapter) updateContractLabelEvents(evMap common.EvMap, kafka *[]common.Message) {
	for idx, contract := range evMap[wasmtypes.EventTypeUpdateContractLabel+"."+wasmtypes.AttributeKeyContractAddr] {
		common.AppendMessage(kafka, "UPDATE_CONTRACT_LABEL", common.JsDict{
			"contract": contract,
			"label":    evMap[wasmtypes.EventTypeUpdateContractLabel+"."+wasmtypes.AttributeKeyNewLabel][idx],
		})
		wa.contractTxs[contract] = false
	}
}

// updateCodeInstantiateConfigEvents handles code access config update events that might be emitted from the transaction
func (wa *WasmAdapter) updateCodeInstantiateConfigEvents(ctx sdk.Context, evMap common.EvMap, kafka *[]common.Message) {
	for _, code := range evMap[wasmtypes.EventTypeUpdateCodeAccessConfig+"."+wasmtypes.AttributeKeyCodeID] {
		codeId := common.Atoui(code)
		codeInfo := wa.wasmKeeper.GetCodeInfo(ctx, codeId)
		addresses := make([]string, 0)
		switch codeInfo.InstantiateConfig.Permission {
		case wasmtypes.AccessTypeAnyOfAddresses:
			addresses = codeInfo.InstantiateConfig.Addresses
		}
		common.AppendMessage(kafka, "UPDATE_CODE_INSTANTIATE_CONFIG", common.JsDict{
			"id":                       codeId,
			"access_config_permission": codeInfo.InstantiateConfig.Permission.String(),
			"access_config_addresses":  addresses,
		})
	}
}
