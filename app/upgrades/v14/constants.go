package v14

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/osmosis-labs/osmosis/v16/app/upgrades"
	downtimetypes "github.com/osmosis-labs/osmosis/v16/x/downtime-detector/types"
	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
)

// UpgradeName defines the on-chain upgrade name for the Osmosis v14 upgrade.
const UpgradeName = "v14"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{downtimetypes.StoreKey, ibchookstypes.StoreKey},
		Deleted: []string{},
	},
}
