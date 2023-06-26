package chain

import (
	"fmt"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	appparams "github.com/osmosis-labs/osmosis/v16/app/params"
	"github.com/osmosis-labs/osmosis/v16/tests/e2e/configurer/config"

	"github.com/osmosis-labs/osmosis/v16/tests/e2e/containers"
	"github.com/osmosis-labs/osmosis/v16/tests/e2e/initialization"
)

type Config struct {
	initialization.ChainMeta

	ValidatorInitConfigs []*initialization.NodeConfig
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	VotingPeriod          float32
	ExpeditedVotingPeriod float32
	// upgrade proposal height for chain.
	UpgradePropHeight    int64
	LatestProposalNumber int
	LatestLockNumber     int
	NodeConfigs          []*NodeConfig
	NodeTempConfigs      []*NodeConfig

	LatestCodeId int

	t                *testing.T
	containerManager *containers.Manager
}

const (
	// defaultNodeIndex to use for querying and executing transactions.
	// It is used when we are indifferent about the node we are working with.
	defaultNodeIndex = 0
	// waitUntilRepeatPauseTime is the time to wait between each check of the node status.
	waitUntilRepeatPauseTime = 2 * time.Second
	// waitUntilrepeatMax is the maximum number of times to repeat the wait until condition.
	waitUntilrepeatMax = 60

	proposalStatusPassed = "PROPOSAL_STATUS_PASSED"
)

func New(t *testing.T, containerManager *containers.Manager, id string, initValidatorConfigs []*initialization.NodeConfig) *Config {
	t.Helper()
	numVal := float32(len(initValidatorConfigs))
	return &Config{
		ChainMeta: initialization.ChainMeta{
			Id: id,
		},
		ValidatorInitConfigs:  initValidatorConfigs,
		VotingPeriod:          config.PropDepositBlocks + numVal*config.PropVoteBlocks + config.PropBufferBlocks,
		ExpeditedVotingPeriod: config.PropDepositBlocks + numVal*config.PropVoteBlocks + config.PropBufferBlocks - 2,
		t:                     t,
		containerManager:      containerManager,
	}
}

// CreateNode returns new initialized NodeConfig.
func (c *Config) CreateNode(initNode *initialization.Node) *NodeConfig {
	nodeConfig := &NodeConfig{
		Node:             *initNode,
		chainId:          c.Id,
		containerManager: c.containerManager,
		t:                c.t,
	}
	c.NodeConfigs = append(c.NodeConfigs, nodeConfig)
	return nodeConfig
}

// CreateNodeTemp returns new initialized NodeConfig and appends it to a separate list of temporary nodes.
func (c *Config) CreateNodeTemp(initNode *initialization.Node) *NodeConfig {
	nodeConfig := &NodeConfig{
		Node:             *initNode,
		chainId:          c.Id,
		containerManager: c.containerManager,
		t:                c.t,
	}
	c.NodeTempConfigs = append(c.NodeTempConfigs, nodeConfig)
	return nodeConfig
}

// RemoveNode removes node and stops it from running.
func (c *Config) RemoveNode(nodeName string) error {
	for i, node := range c.NodeConfigs {
		if node.Name == nodeName {
			c.NodeConfigs = append(c.NodeConfigs[:i], c.NodeConfigs[i+1:]...)
			return node.Stop()
		}
	}
	return fmt.Errorf("node %s not found", nodeName)
}

// RemoveTempNode removes temporary node and stops it from running.
func (c *Config) RemoveTempNode(nodeName string) error {
	for i, node := range c.NodeTempConfigs {
		if node.Name == nodeName {
			c.NodeTempConfigs = append(c.NodeTempConfigs[:i], c.NodeTempConfigs[i+1:]...)
			return node.Stop()
		}
	}
	return fmt.Errorf("node %s not found", nodeName)
}

// WaitUntilEpoch waits for the chain to reach the specified epoch.
func (c *Config) WaitUntilEpoch(epoch int64, epochIdentifier string) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	node.WaitUntil(func(_ coretypes.SyncInfo) bool {
		newEpochNumber := node.QueryCurrentEpoch(epochIdentifier)
		c.t.Logf("current epoch number is (%d), waiting to reach (%d)", newEpochNumber, epoch)
		return newEpochNumber >= epoch
	})
}

// WaitForNumEpochs waits for the chain to to go through a given number of epochs.
func (c *Config) WaitForNumEpochs(epochsToWait int64, epochIdentifier string) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	oldEpochNumber := node.QueryCurrentEpoch(epochIdentifier)
	c.WaitUntilEpoch(oldEpochNumber+epochsToWait, epochIdentifier)
}

// WaitUntilHeight waits for all validators to reach the specified height at the minimum.
// returns error, if any.
func (c *Config) WaitUntilHeight(height int64) {
	// Ensure the nodes are making progress.
	doneCondition := func(syncInfo coretypes.SyncInfo) bool {
		curHeight := syncInfo.LatestBlockHeight

		if curHeight < height {
			c.t.Logf("current block height is %d, waiting to reach: %d", curHeight, height)
			return false
		}

		return !syncInfo.CatchingUp
	}

	for _, node := range c.NodeConfigs {
		c.t.Logf("node container: %s, waiting to reach height %d", node.Name, height)
		node.WaitUntil(doneCondition)
	}
}

// WaitForNumHeights waits for all nodes to go through a given number of heights.
func (c *Config) WaitForNumHeights(heightsToWait int64) {
	node, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	currentHeight, err := node.QueryCurrentHeight()
	require.NoError(c.t, err)
	c.WaitUntilHeight(currentHeight + heightsToWait)
}

func (c *Config) EnableSuperfluidAsset(denom string) {
	chain, err := c.GetDefaultNode()
	require.NoError(c.t, err)
	chain.SubmitSuperfluidProposal(denom, sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)))
	c.LatestProposalNumber += 1
	propNumber := c.LatestProposalNumber
	chain.DepositProposal(propNumber, false)
	var wg sync.WaitGroup

	for _, n := range c.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()
}

func (c *Config) LockAndAddToExistingLock(amount sdk.Int, denom, lockupWalletAddr, lockupWalletSuperfluidAddr string) {
	chain, err := c.GetDefaultNode()
	require.NoError(c.t, err)

	// lock tokens
	chain.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletAddr)
	c.LatestLockNumber += 1
	fmt.Println("lock number: ", c.LatestLockNumber)
	// add to existing lock
	chain.AddToExistingLock(amount, denom, "240s", lockupWalletAddr)

	// superfluid lock tokens
	chain.LockTokens(fmt.Sprintf("%v%s", amount, denom), "240s", lockupWalletSuperfluidAddr)
	c.LatestLockNumber += 1
	fmt.Println("lock number: ", c.LatestLockNumber)
	chain.SuperfluidDelegate(c.LatestLockNumber, c.NodeConfigs[1].OperatorAddress, lockupWalletSuperfluidAddr)
	// add to existing lock
	chain.AddToExistingLock(amount, denom, "240s", lockupWalletSuperfluidAddr)
}

// GetDefaultNode returns the default node of the chain.
// The default node is the first one created. Returns error if no
// ndoes created.
func (c *Config) GetDefaultNode() (*NodeConfig, error) {
	return c.getNodeAtIndex(defaultNodeIndex)
}

// GetPersistentPeers returns persistent peers from every node
// associated with a chain.
func (c *Config) GetPersistentPeers() []string {
	peers := make([]string, len(c.NodeConfigs))
	for i, node := range c.NodeConfigs {
		peers[i] = node.PeerId
	}
	return peers
}

// Returns the nodeIndex'th node on the chain
func (c *Config) GetNodeAtIndex(nodeIndex int) (*NodeConfig, error) {
	return c.getNodeAtIndex(nodeIndex)
}

func (c *Config) getNodeAtIndex(nodeIndex int) (*NodeConfig, error) {
	if nodeIndex > len(c.NodeConfigs) {
		return nil, fmt.Errorf("node index (%d) is greter than the number of nodes available (%d)", nodeIndex, len(c.NodeConfigs))
	}
	return c.NodeConfigs[nodeIndex], nil
}

// func (c *Config) SubmitParamChangeProposal(subspace, key string, value []byte) error {
// 	proposal := paramsutils.ParamChangeProposalJSON{
// 		Title:       "Param Change",
// 		Description: fmt.Sprintf("Changing the %s param", key),
// 		Changes: paramsutils.ParamChangesJSON{
// 			paramsutils.ParamChangeJSON{
// 				Subspace: subspace,
// 				Key:      key,
// 				Value:    value,
// 			},
// 		},
// 		Deposit: "625000000uosmo",
// 	}
// 	proposalJson, err := json.Marshal(proposal)
// 	if err != nil {
// 		return err
// 	}

// 	node, err := c.GetDefaultNode()
// 	if err != nil {
// 		return err
// 	}
// 	propNumber := node.SubmitParamChangeProposal(string(proposalJson), initialization.ValidatorWalletName)
// 	c.LatestProposalNumber += 1

// 	var wg sync.WaitGroup

// 	for _, n := range c.NodeConfigs {
// 		wg.Add(1)
// 		go func(nodeConfig *NodeConfig) {
// 			defer wg.Done()
// 			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
// 		}(n)
// 	}

// 	wg.Wait()

// 	require.Eventually(c.t, func() bool {
// 		status, err := node.QueryPropStatus(propNumber)
// 		if err != nil {
// 			return false
// 		}
// 		return status == proposalStatusPassed
// 	}, time.Minute*30, time.Millisecond*500)
// 	return nil
// }

func (c *Config) SubmitCreateConcentratedPoolProposal() (uint64, error) {
	node, err := c.GetDefaultNode()
	if err != nil {
		return 0, err
	}

	propId := node.SubmitCreateConcentratedPoolProposal(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(config.InitialMinDeposit)))
	c.LatestProposalNumber += 1

	propNumber := propId
	node.DepositProposal(propNumber, false)

	var wg sync.WaitGroup

	for _, n := range c.NodeConfigs {
		wg.Add(1)
		go func(nodeConfig *NodeConfig) {
			defer wg.Done()
			nodeConfig.VoteYesProposal(initialization.ValidatorWalletName, propNumber)
		}(n)
	}

	wg.Wait()

	require.Eventually(c.t, func() bool {
		status, err := node.QueryPropStatus(propNumber)
		if err != nil {
			return false
		}
		return status == proposalStatusPassed
	}, time.Second*30, time.Millisecond*500)
	poolId := node.QueryNumPools()
	return poolId, nil
}

// func (c *Config) SetupRateLimiting(paths, gov_addr string) (string, error) {
// 	node, err := c.GetDefaultNode()
// 	if err != nil {
// 		return "", err
// 	}

// 	// copy the contract from x/rate-limit/testdata/
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		return "", err
// 	}
// 	// go up two levels
// 	projectDir := filepath.Dir(filepath.Dir(wd))
// 	fmt.Println(wd, projectDir)
// 	_, err = util.CopyFile(projectDir+"/x/ibc-rate-limit/bytecode/rate_limiter.wasm", wd+"/scripts/rate_limiter.wasm")
// 	if err != nil {
// 		return "", err
// 	}

// 	codeId := node.StoreWasmCode("rate_limiter.wasm", initialization.ValidatorWalletName)
// 	c.LatestCodeId = int(node.QueryLatestWasmCodeID())
// 	node.InstantiateWasmContract(
// 		strconv.Itoa(codeId),
// 		fmt.Sprintf(`{"gov_module": "%s", "ibc_module": "%s", "paths": [%s] }`, gov_addr, node.PublicAddress, paths),
// 		initialization.ValidatorWalletName)

// 	contracts, err := node.QueryContractsFromId(codeId)
// 	if err != nil {
// 		return "", err
// 	}

// 	contract := contracts[len(contracts)-1]

// 	err = c.SubmitParamChangeProposal(
// 		ibcratelimittypes.ModuleName,
// 		string(ibcratelimittypes.KeyContractAddress),
// 		[]byte(fmt.Sprintf(`"%s"`, contract)),
// 	)
// 	if err != nil {
// 		return "", err
// 	}
// 	require.Eventually(c.t, func() bool {
// 		val := node.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
// 		return strings.Contains(val, contract)
// 	}, time.Second*30, time.Millisecond*500)
// 	fmt.Println("contract address set to", contract)
// 	return contract, nil
// }
