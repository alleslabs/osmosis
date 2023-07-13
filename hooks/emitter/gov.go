package emitter

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/osmosis-labs/osmosis/v16/app/params"
	"github.com/osmosis-labs/osmosis/v16/hooks/common"
)

var (
	StatusInactive         = 6
	_              Adapter = &GovAdapter{}
)

// GovAdapter defines a struct containing the required keeper to process the x/gov hook. It implements Adapter interface.
type GovAdapter struct {
	keeper *govkeeper.Keeper
}

// NewGovAdapter creates a new GovAdapter instance that will be added to the emitter hook adapters.
func NewGovAdapter(keeper *govkeeper.Keeper) *GovAdapter {
	return &GovAdapter{
		keeper: keeper,
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
func (ga *GovAdapter) HandleMsgEvents(ctx sdk.Context, _ []byte, msg sdk.Msg, evMap common.EvMap, detail common.JsDict, kafka *[]common.Message) {
	var submitProposalId uint64
	if rawIds, ok := evMap[types.EventTypeSubmitProposal+"."+types.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			submitProposalId = common.Atoui(rawId)
			proposal, _ := ga.keeper.GetProposal(ctx, submitProposalId)
			content := proposal.GetContent()
			common.AppendMessage(kafka, "NEW_PROPOSAL", common.JsDict{
				"id":               submitProposalId,
				"proposer":         msg.GetSigners()[0],
				"type":             content.ProposalType(),
				"title":            content.GetTitle(),
				"description":      content.GetDescription(),
				"proposal_route":   content.ProposalRoute(),
				"status":           int(proposal.Status),
				"submit_time":      proposal.SubmitTime.UnixNano(),
				"deposit_end_time": proposal.DepositEndTime.UnixNano(),
				"voting_time":      proposal.VotingStartTime.UnixNano(),
				"voting_end_time":  proposal.VotingEndTime.UnixNano(),
				"content":          content,
				"is_expedited":     proposal.IsExpedited,
				"resolved_height":  nil,
			})
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              submitProposalId,
				"status":          int(proposal.Status),
				"voting_time":     proposal.VotingStartTime.UnixNano(),
				"voting_end_time": proposal.VotingEndTime.UnixNano(),
			})
		}
	}

	if rawIds, ok := evMap[types.EventTypeProposalDeposit+"."+types.AttributeKeyVotingPeriodStart]; ok {
		for _, rawId := range rawIds {
			id := common.Atoui(rawId)
			proposal, _ := ga.keeper.GetProposal(ctx, id)
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              id,
				"status":          int(proposal.Status),
				"voting_time":     proposal.VotingStartTime.UnixNano(),
				"voting_end_time": proposal.VotingEndTime.UnixNano(),
			})
		}
	}

	switch msg := msg.(type) {
	case *types.MsgSubmitProposal:
		detail["proposal_id"] = submitProposalId
	case *types.MsgDeposit:
		proposal, _ := ga.keeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetTitle()
	case *types.MsgVote:
		proposal, _ := ga.keeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetTitle()
	case *types.MsgVoteWeighted:
		proposal, _ := ga.keeper.GetProposal(ctx, msg.ProposalId)
		detail["title"] = proposal.GetTitle()
	}
}

// PostDeliverTx does nothing since no action is required after the transaction has been processed by the hook.
func (ga *GovAdapter) PostDeliverTx(_ sdk.Context, _ []byte, _ common.JsDict, _ *[]common.Message) {
}

// AfterEndBlock checks for ActiveProposal or InactiveProposal events and update proposals accordingly.
func (ga *GovAdapter) AfterEndBlock(ctx sdk.Context, _ abci.RequestEndBlock, evMap common.EvMap, kafka *[]common.Message) {
	if rawIds, ok := evMap[types.EventTypeActiveProposal+"."+types.AttributeKeyProposalID]; ok {
		for idx, rawId := range rawIds {
			id := common.Atoui(rawId)
			proposal, _ := ga.keeper.GetProposal(ctx, id)
			if evMap[types.EventTypeActiveProposal+"."+types.AttributeKeyProposalResult][idx] == types.AttributeValueExpeditedProposalRejected {
				common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
					"id":              id,
					"is_expedited":    proposal.IsExpedited,
					"voting_end_time": proposal.VotingEndTime.UnixNano(),
				})
			} else {
				common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
					"id":              id,
					"status":          int(proposal.Status),
					"resolved_height": ctx.BlockHeight(),
				})
			}
		}
	}

	if rawIds, ok := evMap[types.EventTypeInactiveProposal+"."+types.AttributeKeyProposalID]; ok {
		for _, rawId := range rawIds {
			common.AppendMessage(kafka, "UPDATE_PROPOSAL", common.JsDict{
				"id":              common.Atoi(rawId),
				"status":          StatusInactive,
				"resolved_height": ctx.BlockHeight(),
			})
		}
	}
}
