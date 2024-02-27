package emitter

import (
	"encoding/json"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/osmosis-labs/osmosis/v23/app/params"
	"github.com/osmosis-labs/osmosis/v23/hooks/common"
)

var (
	StatusInactive         = 6
	_              Adapter = &GovAdapter{}
)

// GovAdapter defines a struct containing the required keeper to process the x/gov hook. It implements Adapter interface.
type GovAdapter struct {
	govKeeper     *govkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper

	voteInBlock map[uint64]bool
}

// NewGovAdapter creates a new GovAdapter instance that will be added to the emitter hook adapters.
func NewGovAdapter(govKeeper *govkeeper.Keeper, stakingKeeper *stakingkeeper.Keeper) *GovAdapter {
	return &GovAdapter{
		govKeeper:     govKeeper,
		stakingKeeper: stakingKeeper,
		voteInBlock:   make(map[uint64]bool),
	}
}

// AfterInitChain does nothing since no action is required in the InitChainer.
func (ga *GovAdapter) AfterInitChain(_ sdk.Context, _ params.EncodingConfig, _ map[string]json.RawMessage, _ *[]common.Message) {
}

// AfterBeginBlock does nothing since no action is required in the BeginBlocker.
func (ga *GovAdapter) AfterBeginBlock(_ sdk.Context, _ abci.RequestBeginBlock, _ common.EvMap, _ *[]common.Message) {
}

// PreDeliverTx does nothing since no action is required before processing each transaction.
func (ga *GovAdapter) PreDeliverTx() {
}

// CheckMsg does nothing since no message check is required for governance module.
func (ga *GovAdapter) CheckMsg(_ sdk.Context, _ sdk.Msg) {
}

// HandleMsgEvents checks for SubmitProposal or ProposalDeposit events and process new proposals in the current transaction.
func (ga *GovAdapter) HandleMsgEvents(ctx sdk.Context, txHash []byte, msg sdk.Msg, evMap common.EvMap, detail common.JsDict, kafka *[]common.Message) {
	var submitProposalId uint64
	if rawIds, ok := evMap[govtypes.EventTypeSubmitProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			submitProposalId = common.Atoui(rawId)
			proposal, _ := ga.govKeeper.GetProposal(ctx, submitProposalId)

			proposalMsgs := make([]common.JsDict, 0)
			types := make([]string, 0)

			rawProposalMsgs, _ := proposal.GetMsgs()
			for _, proposalMsg := range rawProposalMsgs {
				var proposalMsgJsDict common.JsDict
				proposalMsgJson, err := codec.ProtoMarshalJSON(proposalMsg, nil)
				if err != nil {
					panic(err)
				}
				err = json.Unmarshal(proposalMsgJson, &proposalMsgJsDict)
				proposalMsgJsDict["@type"] = sdk.MsgTypeURL(proposalMsg)
				proposalMsgs = append(proposalMsgs, proposalMsgJsDict)

				switch typedMsg := proposalMsg.(type) {
				case *govv1types.MsgExecLegacyContent:
					content, _ := govv1types.LegacyContentFromMessage(typedMsg)
					types = append(types, content.ProposalType())
				default:
					types = append(types, sdk.MsgTypeURL(typedMsg))
				}
			}
			newProposal := common.JsDict{
				"id":               submitProposalId,
				"proposer":         msg.GetSigners()[0],
				"title":            proposal.GetTitle(),
				"description":      proposal.GetSummary(),
				"status":           int(proposal.Status),
				"submit_time":      proposal.SubmitTime.UnixNano(),
				"deposit_end_time": proposal.DepositEndTime.UnixNano(),
				"total_deposit":    proposal.TotalDeposit,
				"yes":              0,
				"no":               0,
				"abstain":          0,
				"no_with_veto":     0,
				"is_expedited":     false,
				"resolved_height":  nil,
				"types":            types,
				"messages":         proposalMsgs,
				"created_tx":       txHash,
				"created_height":   ctx.BlockHeight(),
				"metadata":         proposal.GetMetadata(),
			}

			if len(rawProposalMsgs) > 0 {
				switch proposalMsg := rawProposalMsgs[0].(type) {
				case *govv1types.MsgExecLegacyContent:
					content, _ := govv1types.LegacyContentFromMessage(proposalMsg)
					newProposal["type"] = content.ProposalType()
					newProposal["proposal_route"] = content.ProposalRoute()
					newProposal["content"] = content
					newProposal["version"] = "v1beta1"
				default:
					metadata := proposal.GetMetadata()
					newProposal["type"] = sdk.MsgTypeURL(proposalMsg)
					newProposal["proposal_route"] = govtypes.RouterKey
					newProposal["content"] = common.JsDict{"messages": proposalMsgs, "metadata": metadata}
					newProposal["version"] = "v1"
				}
			} else {
				newProposal["type"] = ""
				newProposal["proposal_route"] = govtypes.RouterKey
				newProposal["content"] = common.JsDict{"messages": []common.JsDict{}, "metadata": proposal.GetMetadata()}
				newProposal["version"] = "v1"
			}

			common.AppendMessage(kafka, "NEW_PROPOSAL", newProposal)
		}
	}

	if rawIds, ok := evMap[govtypes.EventTypeProposalDeposit+"."+govtypes.AttributeKeyProposalID]; ok {
		rawAmounts, ok := evMap[govtypes.EventTypeProposalDeposit+"."+sdk.AttributeKeyAmount]
		if !ok {
			panic("unable to get proposal deposit amount")
		}
		if len(rawIds) != len(rawAmounts) {
			panic("deposit proposal ids and amounts doesn't match")
		}

		for idx, _ := range rawIds {
			id := common.Atoui(rawIds[idx])
			amount, err := sdk.ParseCoinsNormalized(rawAmounts[idx])
			if err != nil {
				panic(err)
			}

			common.AppendMessage(kafka, "NEW_PROPOSAL_DEPOSIT", common.JsDict{
				"proposal_id": id,
				"tx_hash":     txHash,
				"depositor":   msg.GetSigners()[0],
				"amount":      amount,
			})

			proposal, _ := ga.govKeeper.GetProposal(ctx, id)
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":            id,
				"total_deposit": proposal.TotalDeposit,
			})
		}
	}

	if rawIds, ok := evMap[govtypes.EventTypeProposalDeposit+"."+govtypes.AttributeKeyVotingPeriodStart]; ok {
		for _, rawId := range rawIds {
			id := common.Atoui(rawId)
			proposal, _ := ga.govKeeper.GetProposal(ctx, id)
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              id,
				"status":          int(proposal.Status),
				"voting_time":     proposal.VotingStartTime.UnixNano(),
				"voting_end_time": proposal.VotingEndTime.UnixNano(),
			})
		}
	}

	if rawIds, ok := evMap[govtypes.EventTypeSubmitProposal+"."+govtypes.AttributeKeyVotingPeriodStart]; ok {
		for _, rawId := range rawIds {
			id := common.Atoui(rawId)
			proposal, _ := ga.govKeeper.GetProposal(ctx, id)
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              id,
				"status":          int(proposal.Status),
				"voting_time":     proposal.VotingStartTime.UnixNano(),
				"voting_end_time": proposal.VotingEndTime.UnixNano(),
			})
		}
	}

	if rawIds, ok := evMap[govtypes.EventTypeProposalVote+"."+govtypes.AttributeKeyProposalID]; ok {
		rawOptions, ok := evMap[govtypes.EventTypeProposalVote+"."+govtypes.AttributeKeyOption]
		if !ok {
			panic("unable to get proposal vote options")
		}
		if len(rawIds) != len(rawOptions) {
			panic("voting proposal ids and options amount doesn't match")
		}

		for idx, _ := range rawIds {
			id := common.Atoui(rawIds[idx])
			votes := map[govv1types.VoteOption]string{govv1types.OptionYes: "0", govv1types.OptionAbstain: "0", govv1types.OptionNo: "0", govv1types.OptionNoWithVeto: "0"}
			var options govv1types.WeightedVoteOptions
			for _, rawOption := range strings.Split(rawOptions[idx], "\n") {
				fields := strings.Split(rawOption, " ")
				option, err := govv1types.VoteOptionFromString(strings.Split(fields[0], ":")[1])
				if err != nil {
					panic(err)
				}
				if len(fields) < 2 {
					panic(fmt.Errorf("weight field does not exist for %s option", fields[0]))
				}
				weight := strings.Trim(strings.Split(fields[1], ":")[1], "\"")
				if err != nil {
					panic(err)
				}
				options = append(options, &govv1types.WeightedVoteOption{Option: option, Weight: weight})
				votes[option] = weight
			}

			rawVoter := evMap[govtypes.EventTypeProposalVote+"."+govtypes.AttributeKeyVoter][idx]
			voter, _ := sdk.AccAddressFromBech32(rawVoter)
			newVote := common.JsDict{
				"proposal_id":      id,
				"tx_hash":          txHash,
				"voter":            voter,
				"is_vote_weighted": len(options) > 1,
				"is_validator":     false,
				"yes":              votes[govv1types.OptionYes],
				"no":               votes[govv1types.OptionNo],
				"abstain":          votes[govv1types.OptionAbstain],
				"no_with_veto":     votes[govv1types.OptionNoWithVeto],
			}

			validator, found := ga.stakingKeeper.GetValidator(ctx, sdk.ValAddress(voter))
			if found {
				newVote["is_validator"] = true
				newVote["validator_address"] = validator.GetOperator().String()
			}

			common.AppendMessage(kafka, "NEW_PROPOSAL_VOTE", newVote)
			ga.voteInBlock[id] = true
		}
	}

	switch msg := msg.(type) {
	case *govv1types.MsgSubmitProposal:
		detail["proposal_id"] = submitProposalId
	case *govv1types.MsgDeposit:
		proposal, _ := ga.govKeeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetMetadata()
	case *govv1types.MsgVote:
		proposal, _ := ga.govKeeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetMetadata()
	case *govv1types.MsgVoteWeighted:
		proposal, _ := ga.govKeeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetMetadata()
	}
}

// PostDeliverTx does nothing since no action is required after the transaction has been processed by the hook.
func (ga *GovAdapter) PostDeliverTx(_ sdk.Context, _ []byte, _ common.JsDict, _ *[]common.Message) {
}

// AfterEndBlock checks for ActiveProposal or InactiveProposal events and update proposals accordingly.
func (ga *GovAdapter) AfterEndBlock(ctx sdk.Context, _ abci.RequestEndBlock, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for idx, rawId := range rawIds {
			id := common.Atoui(rawId)
			proposal, _ := ga.govKeeper.GetProposal(ctx, id)
			if evMap[govtypes.EventTypeActiveProposal+"."+govtypes.AttributeKeyProposalResult][idx] == govtypes.AttributeValueExpeditedProposalRejected {
				common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
					"id":              id,
					"is_expedited":    proposal.Expedited,
					"voting_end_time": proposal.VotingEndTime.UnixNano(),
				})
			} else {
				common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
					"id":              id,
					"status":          int(proposal.Status),
					"yes":             proposal.GetFinalTallyResult().GetYesCount(),
					"no":              proposal.GetFinalTallyResult().GetNoCount(),
					"abstain":         proposal.GetFinalTallyResult().GetAbstainCount(),
					"no_with_veto":    proposal.GetFinalTallyResult().GetNoWithVetoCount(),
					"resolved_height": ctx.BlockHeight(),
				})
			}
		}
	}

	if rawIds, ok := evMap[govtypes.EventTypeInactiveProposal+"."+govtypes.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              common.Atoi(rawId),
				"status":          StatusInactive,
				"resolved_height": ctx.BlockHeight(),
			})
		}
	}
}

//// flushUpdateProposalVotes appends updated governance proposals votes into the provided Kafka messages array.
//func (ga *GovAdapter) flushUpdateProposalVotes(ctx sdk.Context, kafka *[]common.Message) {
//	for proposalId := range ga.voteInBlock {
//		proposal, _ := ga.govKeeper.GetProposal(ctx, proposalId)
//		_, _, tallyResults := ga.govKeeper.Tally(ctx, proposal)
//
//		common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
//			"id":           proposalId,
//			"yes":          tallyResults.GetYesCount(),
//			"no":           tallyResults.GetNoCount(),
//			"abstain":      tallyResults.GetAbstainCount(),
//			"no_with_veto": tallyResults.GetNoWithVetoCount(),
//		})
//	}
//	ga.voteInBlock = make(map[uint64]bool)
//}
