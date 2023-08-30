package common_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/assert"

	"github.com/osmosis-labs/osmosis/v19/app"
	"github.com/osmosis-labs/osmosis/v19/hooks/common"
)

var (
	ALICE      = sdk.AccAddress("01234567890123456789")
	BOB        = sdk.AccAddress("98765432109876543210")
	CAROL      = sdk.AccAddress("01234567899876543210")
	CONTRACT   = common.BuildContractAddressClassic(1, 3)
	FirstCoin  = sdk.NewCoins(sdk.NewCoin("first", sdk.NewInt(1)))
	SecondCoin = sdk.NewCoins(sdk.NewCoin("second", sdk.NewInt(1000)))
)

func GetMsgInterface(msg sdk.Msg) interface{} {
	encoding := app.MakeEncodingConfig()
	bz, err := encoding.Marshaler.MarshalInterfaceJSON(msg)
	fmt.Println(err)
	var jsonMsg interface{}
	_ = json.Unmarshal(bz, &jsonMsg)
	return jsonMsg
}

func TestMsgSend(t *testing.T) {
	accs := make(map[string]bool)
	msg := banktypes.NewMsgSend(ALICE, BOB, FirstCoin)
	common.GetRelatedAccounts(GetMsgInterface(msg), accs)
	assert.Contains(t, accs, ALICE.String())
	assert.Contains(t, accs, BOB.String())
}

func TestFundCom(t *testing.T) {
	accs := make(map[string]bool)
	msg := disttypes.NewMsgFundCommunityPool(FirstCoin, ALICE)
	common.GetRelatedAccounts(GetMsgInterface(msg), accs)
	assert.Contains(t, accs, ALICE.String())
}

func TestMultiSend(t *testing.T) {
	accs := make(map[string]bool)
	inputs := []banktypes.Input{banktypes.NewInput(ALICE, FirstCoin), banktypes.NewInput(BOB, SecondCoin)}
	outputs := []banktypes.Output{
		banktypes.NewOutput(CAROL,
			sdk.Coins{
				sdk.NewCoin("jaja", sdk.NewInt(1)),
				sdk.NewCoin("yoyo", sdk.NewInt(1000)),
			}),
	}

	msg := banktypes.NewMsgMultiSend(inputs, outputs)
	common.GetRelatedAccounts(GetMsgInterface(msg), accs)
	assert.Contains(t, accs, ALICE.String())
	assert.Contains(t, accs, BOB.String())
	assert.Contains(t, accs, CAROL.String())
}

func TestMsgExec(t *testing.T) {
	accs := make(map[string]bool)
	msg := authz.NewMsgExec(ALICE, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("steak", 2)),
			FromAddress: BOB.String(),
			ToAddress:   CAROL.String(),
		},
	})
	sdkMsg := []sdk.Msg{&msg}

	common.GetRelatedAccounts(GetMsgInterface(sdkMsg[0]), accs)
	assert.Contains(t, accs, ALICE.String())
	assert.Contains(t, accs, BOB.String())
	assert.Contains(t, accs, CAROL.String())
}

func TestMsgExecute(t *testing.T) {
	accs := make(map[string]bool)

	msg := types.MsgExecuteContract{
		Sender:   ALICE.String(),
		Contract: CONTRACT.String(),
		Msg:      []byte(`{"release": "osmo15d4apf20449ajvwycq8ruaypt7v6d345z36n9l", "1" : "osmo1zs3qcgu5nsyeq2wnv480uwhqck4jwgza5as3k25wysglggcny0rstfqyz9","key" : {"value":"osmo1mhznfr60vjdp2gejhyv2gax9nvyyzhd3z0qcwseyetkfustjauzqycsy2g","ja":1}}`),
		Funds:    FirstCoin,
	}

	sdkMsg := []sdk.Msg{&msg}
	common.GetRelatedAccounts(GetMsgInterface(sdkMsg[0]), accs)
	assert.Contains(t, accs, ALICE.String())
	assert.Contains(t, accs, CONTRACT.String())
	assert.Contains(t, accs, "osmo15d4apf20449ajvwycq8ruaypt7v6d345z36n9l")
	assert.Contains(t, accs, "osmo1zs3qcgu5nsyeq2wnv480uwhqck4jwgza5as3k25wysglggcny0rstfqyz9")
	assert.Contains(t, accs, "osmo1mhznfr60vjdp2gejhyv2gax9nvyyzhd3z0qcwseyetkfustjauzqycsy2g")
}
