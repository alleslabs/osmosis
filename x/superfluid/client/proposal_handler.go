package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/osmosis-labs/osmosis/v16/x/superfluid/client/cli"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/client/rest"
)

var (
	SetSuperfluidAssetsProposalHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitSetSuperfluidAssetsProposal, rest.ProposalSetSuperfluidAssetsRESTHandler)
	RemoveSuperfluidAssetsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitRemoveSuperfluidAssetsProposal, rest.ProposalRemoveSuperfluidAssetsRESTHandler)
	UpdateUnpoolWhitelistProposalHandler  = govclient.NewProposalHandler(cli.NewCmdUpdateUnpoolWhitelistProposal, rest.ProposalUpdateUnpoolWhitelistProposal)
)
