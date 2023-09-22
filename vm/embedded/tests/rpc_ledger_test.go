package tests

import (
	"encoding/json"
	"fmt"
	"github.com/zenon-network/go-zenon/rpc/api/embedded"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"math/big"
	"testing"
	"time"

	g "github.com/zenon-network/go-zenon/chain/genesis/mock"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/zenon/mock"
)

func TestRPCLedger_PublishRawTX_internal(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()

	simpleSendSetup(t, z)

	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 2, 1)).Equals(t, `
{
	"list": [
		{
			"version": 1,
			"chainIdentifier": 100,
			"blockType": 2,
			"hash": "6e9bf5f7512931a4b74d3d1dd20b0f8105a006b1ae059e1535f935e283f2a66c",
			"previousHash": "598fa623dd308bec7163bb375aa7546ec4aced3b71a1c9278709903e69280dbd",
			"height": 2,
			"momentumAcknowledged": {
				"hash": "0385d849ee33b94c8783288c148e3ae741c2ecec98b08b3f59d6bcc219168fe5",
				"height": 1
			},
			"address": "z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
			"toAddress": "z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
			"amount": "10000000000",
			"tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
			"fromBlockHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"descendantBlocks": [],
			"data": "",
			"fusedPlasma": 21000,
			"difficulty": 0,
			"nonce": "0000000000000000",
			"basePlasma": 21000,
			"usedPlasma": 21000,
			"changesHash": "d3e45796519a8312b6c50f32e49fec272b6ad13f343f1a87dd15f1672059e570",
			"publicKey": "GYyn77OXTL31zPbDBCe/eKir+VCF3hv+LxiOUF3XcJY=",
			"signature": "130sas2Jlmu5AC5SsvJ3I0m31WtvzTKmB3DfoAROQ7kuvx/Hd/g+eZn5rSW5+o5jxV5BJtq1vITs/3lCieGaAw==",
			"token": {
				"name": "Zenon Coin",
				"symbol": "ZNN",
				"domain": "zenon.network",
				"totalSupply": "19500000000000",
				"decimals": 8,
				"owner": "z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg",
				"tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
				"maxSupply": "4611686018427387903",
				"isBurnable": true,
				"isMintable": true,
				"isUtility": true
			},
			"confirmationDetail": {
				"numConfirmations": 2,
				"momentumHeight": 2,
				"momentumHash": "ea202e600eb999ad1bb46788a46c9bebc7c6795c772cbb1f5a262a29a77da740",
				"momentumTimestamp": 1000000010
			},
			"pairedAccountBlock": {
				"version": 1,
				"chainIdentifier": 100,
				"blockType": 3,
				"hash": "f845e19928c2452b96c88ff49b60d5e3fa7632a86006a951f80fe8a22dbeb810",
				"previousHash": "57b6b7c6edb82b38ec4c992d99c84bf8016f03bf0727ff9daa811d2e862fa77a",
				"height": 2,
				"momentumAcknowledged": {
					"hash": "ea202e600eb999ad1bb46788a46c9bebc7c6795c772cbb1f5a262a29a77da740",
					"height": 2
				},
				"address": "z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
				"toAddress": "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
				"amount": "0",
				"tokenStandard": "zts1qqqqqqqqqqqqqqqqtq587y",
				"fromBlockHash": "6e9bf5f7512931a4b74d3d1dd20b0f8105a006b1ae059e1535f935e283f2a66c",
				"descendantBlocks": [],
				"data": "",
				"fusedPlasma": 21000,
				"difficulty": 0,
				"nonce": "0000000000000000",
				"basePlasma": 21000,
				"usedPlasma": 21000,
				"changesHash": "fdc4b5908226e07dcc97f7ecfba7e2feacce4c93d5e9e9e44ed4fcf870b53859",
				"publicKey": "tUJu3P7Drp25XP662lIjyFlFpvj8bWUpyC+0y5YTzXM=",
				"signature": "IBFUD+6F4sZ5BTS1IzZH1IT/QVc8laskvtoieKQbOM6BGEVy4dstAlznSNHrblBnorKUDofuR7ekjfPem2wGBQ==",
				"token": null,
				"confirmationDetail": {
					"numConfirmations": 1,
					"momentumHeight": 3,
					"momentumHash": "7c8f4900aea3b2b2c91fb26ce0d4269f92c7ff80d7a0e8bcdb1cf3b8ed7411f3",
					"momentumTimestamp": 1000000020
				},
				"pairedAccountBlock": null
			}
		}
	],
	"count": 2,
	"more": false
}`)
	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User2.Address, 2, 1)).Equals(t, `
{
	"list": [
		{
			"version": 1,
			"chainIdentifier": 100,
			"blockType": 3,
			"hash": "f845e19928c2452b96c88ff49b60d5e3fa7632a86006a951f80fe8a22dbeb810",
			"previousHash": "57b6b7c6edb82b38ec4c992d99c84bf8016f03bf0727ff9daa811d2e862fa77a",
			"height": 2,
			"momentumAcknowledged": {
				"hash": "ea202e600eb999ad1bb46788a46c9bebc7c6795c772cbb1f5a262a29a77da740",
				"height": 2
			},
			"address": "z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
			"toAddress": "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
			"amount": "0",
			"tokenStandard": "zts1qqqqqqqqqqqqqqqqtq587y",
			"fromBlockHash": "6e9bf5f7512931a4b74d3d1dd20b0f8105a006b1ae059e1535f935e283f2a66c",
			"descendantBlocks": [],
			"data": "",
			"fusedPlasma": 21000,
			"difficulty": 0,
			"nonce": "0000000000000000",
			"basePlasma": 21000,
			"usedPlasma": 21000,
			"changesHash": "fdc4b5908226e07dcc97f7ecfba7e2feacce4c93d5e9e9e44ed4fcf870b53859",
			"publicKey": "tUJu3P7Drp25XP662lIjyFlFpvj8bWUpyC+0y5YTzXM=",
			"signature": "IBFUD+6F4sZ5BTS1IzZH1IT/QVc8laskvtoieKQbOM6BGEVy4dstAlznSNHrblBnorKUDofuR7ekjfPem2wGBQ==",
			"token": null,
			"confirmationDetail": {
				"numConfirmations": 1,
				"momentumHeight": 3,
				"momentumHash": "7c8f4900aea3b2b2c91fb26ce0d4269f92c7ff80d7a0e8bcdb1cf3b8ed7411f3",
				"momentumTimestamp": 1000000020
			},
			"pairedAccountBlock": {
				"version": 1,
				"chainIdentifier": 100,
				"blockType": 2,
				"hash": "6e9bf5f7512931a4b74d3d1dd20b0f8105a006b1ae059e1535f935e283f2a66c",
				"previousHash": "598fa623dd308bec7163bb375aa7546ec4aced3b71a1c9278709903e69280dbd",
				"height": 2,
				"momentumAcknowledged": {
					"hash": "0385d849ee33b94c8783288c148e3ae741c2ecec98b08b3f59d6bcc219168fe5",
					"height": 1
				},
				"address": "z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
				"toAddress": "z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
				"amount": "10000000000",
				"tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
				"fromBlockHash": "0000000000000000000000000000000000000000000000000000000000000000",
				"descendantBlocks": [],
				"data": "",
				"fusedPlasma": 21000,
				"difficulty": 0,
				"nonce": "0000000000000000",
				"basePlasma": 21000,
				"usedPlasma": 21000,
				"changesHash": "d3e45796519a8312b6c50f32e49fec272b6ad13f343f1a87dd15f1672059e570",
				"publicKey": "GYyn77OXTL31zPbDBCe/eKir+VCF3hv+LxiOUF3XcJY=",
				"signature": "130sas2Jlmu5AC5SsvJ3I0m31WtvzTKmB3DfoAROQ7kuvx/Hd/g+eZn5rSW5+o5jxV5BJtq1vITs/3lCieGaAw==",
				"token": {
					"name": "Zenon Coin",
					"symbol": "ZNN",
					"domain": "zenon.network",
					"totalSupply": "19500000000000",
					"decimals": 8,
					"owner": "z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg",
					"tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
					"maxSupply": "4611686018427387903",
					"isBurnable": true,
					"isMintable": true,
					"isUtility": true
				},
				"confirmationDetail": {
					"numConfirmations": 2,
					"momentumHeight": 2,
					"momentumHash": "ea202e600eb999ad1bb46788a46c9bebc7c6795c772cbb1f5a262a29a77da740",
					"momentumTimestamp": 1000000010
				},
				"pairedAccountBlock": null
			}
		}
	],
	"count": 2,
	"more": false
}`)
}
func TestRPCLedger_PublishRawTransaction(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()

	a := &api.AccountBlock{}
	common.FailIfErr(t, json.Unmarshal([]byte(`
{
  "version": 1,
  "chainIdentifier": 100,
  "blockType": 2,
  "hash": "6e9bf5f7512931a4b74d3d1dd20b0f8105a006b1ae059e1535f935e283f2a66c",
  "previousHash": "598fa623dd308bec7163bb375aa7546ec4aced3b71a1c9278709903e69280dbd",
  "height": 2,
  "momentumAcknowledged": {
    "hash": "0385d849ee33b94c8783288c148e3ae741c2ecec98b08b3f59d6bcc219168fe5",
    "height": 1
  },
  "address": "z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
  "toAddress": "z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
  "amount": "10000000000",
  "tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
  "fromBlockHash": "0000000000000000000000000000000000000000000000000000000000000000",
  "data": "",
  "fusedPlasma": 21000,
  "difficulty": 0,
  "nonce": "0000000000000000",
  "publicKey": "GYyn77OXTL31zPbDBCe/eKir+VCF3hv+LxiOUF3XcJY=",
  "signature": "130sas2Jlmu5AC5SsvJ3I0m31WtvzTKmB3DfoAROQ7kuvx/Hd/g+eZn5rSW5+o5jxV5BJtq1vITs/3lCieGaAw=="
}`), a))
	common.FailIfErr(t, ledgerApi.PublishRawTransaction(a))
}

func ExpectGetFrontierAccountBlock(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)
	common.Json(ledgerApi.GetFrontierAccountBlock(g.User1.Address)).SubJson(&Height{}).Equals(t, `
{
	"height": 11
}`)
}
func ExpectGetAccountBlocksByHeight(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)
	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 3, 2)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 3
		},
		{
			"height": 4
		}
	]
}`)
	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 1, 5)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 1
		},
		{
			"height": 2
		},
		{
			"height": 3
		},
		{
			"height": 4
		},
		{
			"height": 5
		}
	]
}`)
	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 20, 5)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": []
}`)
	common.Json(ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 10, 5)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 10
		},
		{
			"height": 11
		}
	]
}`)
}
func ExpectGetAccountBlockByHash(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)

	blocks, err := ledgerApi.GetAccountBlocksByHeight(g.User1.Address, 1, 10)
	common.FailIfErr(t, err)
	common.Json(ledgerApi.GetAccountBlockByHash(blocks.List[0].Hash)).SubJson(&Height{}).Equals(t, `
{
	"height": 1
}`)
	common.Json(ledgerApi.GetAccountBlockByHash(blocks.List[5].Hash)).SubJson(&Height{}).Equals(t, `
{
	"height": 6
}`)
	common.Json(ledgerApi.GetAccountBlockByHash(types.NewHash([]byte{'1'}))).SubJson(&Height{}).Equals(t, `null`)
}
func ExpectGetAccountBlocksByPage(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)

	common.Json(ledgerApi.GetAccountBlocksByPage(g.User1.Address, 0, 2)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 11
		},
		{
			"height": 10
		}
	]
}`)
	common.Json(ledgerApi.GetAccountBlocksByPage(g.User1.Address, 2, 2)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 7
		},
		{
			"height": 6
		}
	]
}`)
	common.Json(ledgerApi.GetAccountBlocksByPage(g.User1.Address, 1, 8)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"height": 3
		},
		{
			"height": 2
		},
		{
			"height": 1
		}
	]
}`)
	common.Json(ledgerApi.GetAccountBlocksByPage(g.User1.Address, 2, 8)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 11,
	"list": []
}`)
}
func ExpectGetAccountInfoByAddress(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)

	common.Json(ledgerApi.GetAccountInfoByAddress(g.User1.Address)).Equals(t, `
{
	"address": "z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
	"accountHeight": 11,
	"balanceInfoMap": {
		"zts1qsrxxxxxxxxxxxxxmrhjll": {
			"token": {
				"name": "QuasarCoin",
				"symbol": "QSR",
				"domain": "zenon.network",
				"totalSupply": "180550000000000",
				"decimals": 8,
				"owner": "z1qxemdeddedxstakexxxxxxxxxxxxxxxxjv8v62",
				"tokenStandard": "zts1qsrxxxxxxxxxxxxxmrhjll",
				"maxSupply": "4611686018427387903",
				"isBurnable": true,
				"isMintable": true,
				"isUtility": true
			},
			"balance": "12000000000000"
		},
		"zts1znnxxxxxxxxxxxxx9z4ulx": {
			"token": {
				"name": "Zenon Coin",
				"symbol": "ZNN",
				"domain": "zenon.network",
				"totalSupply": "19500000000000",
				"decimals": 8,
				"owner": "z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg",
				"tokenStandard": "zts1znnxxxxxxxxxxxxx9z4ulx",
				"maxSupply": "4611686018427387903",
				"isBurnable": true,
				"isMintable": true,
				"isUtility": true
			},
			"balance": "1190000000000"
		}
	}
}`)
}
func ExpectGetUnreceivedBlocksByAddress(t *testing.T, z mock.MockZenon) {
	ledgerApi := api.NewLedgerApi(z)

	z.InsertNewMomentum()
	common.Json(ledgerApi.GetUnreceivedBlocksByAddress(g.User2.Address, 0, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 10
		},
		{
			"height": 7
		},
		{
			"height": 3
		},
		{
			"height": 11
		},
		{
			"height": 9
		},
		{
			"height": 2
		},
		{
			"height": 6
		}
	]
}`)
	common.Json(ledgerApi.GetUnreceivedBlocksByAddress(g.User2.Address, 1, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 4
		},
		{
			"height": 8
		},
		{
			"height": 5
		}
	]
}`)
	common.Json(ledgerApi.GetUnreceivedBlocksByAddress(g.User2.Address, 2, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": []
}`)
	autoreceive(t, z, g.User2.Address)
	common.Json(ledgerApi.GetUnreceivedBlocksByAddress(g.User2.Address, 0, 10)).Equals(t, `
{
	"list": [],
	"count": 0,
	"more": false
}`)
}

func TestRPCLedger_UnconfirmedAccountBlocks(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()

	for i := 0; i < 10; i += 1 {
		z.InsertSendBlock(&nom.AccountBlock{
			Address:       g.User1.Address,
			ToAddress:     g.User2.Address,
			TokenStandard: types.ZnnTokenStandard,
			Amount:        big.NewInt(10 * g.Zexp),
		}, nil, mock.SkipVmChanges)
	}

	common.Json(ledgerApi.GetUnconfirmedBlocksByAddress(g.User1.Address, 0, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 2
		},
		{
			"height": 3
		},
		{
			"height": 4
		},
		{
			"height": 5
		},
		{
			"height": 6
		},
		{
			"height": 7
		},
		{
			"height": 8
		}
	]
}`)
	common.Json(ledgerApi.GetUnconfirmedBlocksByAddress(g.User1.Address, 1, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 9
		},
		{
			"height": 10
		},
		{
			"height": 11
		}
	]
}`)
	common.Json(ledgerApi.GetUnconfirmedBlocksByAddress(g.User1.Address, 2, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": []
}`)

	ExpectGetFrontierAccountBlock(t, z)
	ExpectGetAccountBlocksByHeight(t, z)
	//ExpectGetAccountBlockByHash(t, z)
	ExpectGetAccountBlocksByPage(t, z)
	ExpectGetAccountInfoByAddress(t, z)
	ExpectGetUnreceivedBlocksByAddress(t, z)
}
func TestRPCLedger_ConfirmedAccountBlocks(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()

	for i := 0; i < 10; i += 1 {
		z.InsertSendBlock(&nom.AccountBlock{
			Address:       g.User1.Address,
			ToAddress:     g.User2.Address,
			TokenStandard: types.ZnnTokenStandard,
			Amount:        big.NewInt(10 * g.Zexp),
		}, nil, mock.SkipVmChanges)
	}
	z.InsertNewMomentum()

	common.Json(ledgerApi.GetUnconfirmedBlocksByAddress(g.User1.Address, 0, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 0,
	"list": []
}`)

	ExpectGetFrontierAccountBlock(t, z)
	ExpectGetAccountBlocksByHeight(t, z)
	ExpectGetAccountBlockByHash(t, z)
	ExpectGetAccountBlocksByPage(t, z)
	ExpectGetAccountInfoByAddress(t, z)
	ExpectGetUnreceivedBlocksByAddress(t, z)
}

func TestRPCLedger_GetFrontierMomentum(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	common.Json(ledgerApi.GetFrontierMomentum()).SubJson(&Height{}).Equals(t, `
{
	"height": 10
}`)
}
func TestRPCLedger_GetMomentumBeforeTime(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	momentums, err := ledgerApi.GetMomentumsByHeight(1, 1)
	common.DealWithErr(err)
	genesis := momentums.List[0]

	common.Json(ledgerApi.GetMomentumBeforeTime(genesis.Timestamp.Add(time.Second*10*5+time.Second).Unix())).SubJson(&Height{}).Equals(t, `
{
	"height": 6
}`)
	common.Json(ledgerApi.GetMomentumBeforeTime(genesis.Timestamp.Add(time.Hour).Unix())).SubJson(&Height{}).Equals(t, `
{
	"height": 10
}`)
	common.Json(ledgerApi.GetMomentumBeforeTime(genesis.Timestamp.Add(-time.Hour).Unix())).SubJson(&Height{}).Equals(t, `null`)
}
func TestRPCLedger_GetMomentumByHash(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	momentum, err := ledgerApi.GetFrontierMomentum()
	common.DealWithErr(err)
	common.Json(ledgerApi.GetMomentumByHash(momentum.Hash)).SubJson(&Height{}).Equals(t, `
{
	"height": 10
}`)
}
func TestRPCLedger_GetMomentumsByHeight(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	common.Json(ledgerApi.GetMomentumsByHeight(2, 3)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 2
		},
		{
			"height": 3
		},
		{
			"height": 4
		}
	]
}`)
	common.Json(ledgerApi.GetMomentumsByHeight(8, 5)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 8
		},
		{
			"height": 9
		},
		{
			"height": 10
		}
	]
}`)
	common.Json(ledgerApi.GetMomentumsByHeight(15, 5)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": []
}`)
}
func TestRPCLedger_GetMomentumsByPage(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	common.Json(ledgerApi.GetMomentumsByPage(0, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 10
		},
		{
			"height": 9
		},
		{
			"height": 8
		},
		{
			"height": 7
		},
		{
			"height": 6
		},
		{
			"height": 5
		},
		{
			"height": 4
		}
	]
}`)
	common.Json(ledgerApi.GetMomentumsByPage(1, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": [
		{
			"height": 3
		},
		{
			"height": 2
		},
		{
			"height": 1
		}
	]
}`)
	common.Json(ledgerApi.GetMomentumsByPage(2, 7)).SubJson(ListOfHeight()).Equals(t, `
{
	"count": 10,
	"list": []
}`)
}
func TestRPCLedger_GetDetailedMomentumsByHeight(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)
	z.InsertNewMomentum()
	common.Json(ledgerApi.GetDetailedMomentumsByHeight(1, 3)).SubJson(ListOf(func() interface{} {
		return new(struct {
			AccountBlocks *listToCount `json:"blocks"`
			Momentum      *struct {
				Height uint64 `json:"height"`
			} `json:"momentum"`
		})
	})).Equals(t, `
{
	"count": 11,
	"list": [
		{
			"blocks": 18,
			"momentum": {
				"height": 1
			}
		},
		{
			"blocks": 0,
			"momentum": {
				"height": 2
			}
		},
		{
			"blocks": 0,
			"momentum": {
				"height": 3
			}
		}
	]
}`)
}
func TestRPCLedger_Errors(t *testing.T) {
	z := mock.NewMockZenon(t)
	ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()

	common.Json(ledgerApi.GetDetailedMomentumsByHeight(0, 3)).Error(t, api.ErrHeightParamIsZero)
	common.Json(ledgerApi.GetDetailedMomentumsByHeight(1, 1234)).Error(t, api.ErrCountParamTooBig)
	common.Json(ledgerApi.GetAccountBlocksByPage(types.ZeroAddress, 0, 1234)).Error(t, api.ErrPageSizeParamTooBig)
}

func activateFee(z mock.MockZenon) {
	z.InsertSendBlock(&nom.AccountBlock{
		Address:   g.Spork.Address,
		ToAddress: types.SporkContract,
		Data: definition.ABISpork.PackMethodPanic(definition.SporkCreateMethodName,
			"spork-fee",              // name
			"activate spork for fee", // description
		),
	}, nil, mock.SkipVmChanges)
	z.InsertNewMomentum()

	sporkAPI := embedded.NewSporkApi(z)
	sporkList, _ := sporkAPI.GetAll(0, 10)
	id := sporkList.List[0].Id

	z.InsertSendBlock(&nom.AccountBlock{
		Address:   g.Spork.Address,
		ToAddress: types.SporkContract,
		Data: definition.ABISpork.PackMethodPanic(definition.SporkActivateMethodName,
			id, // id
		),
	}, nil, mock.SkipVmChanges)
	z.InsertNewMomentum()
	types.FeeSpork.SporkId = id
	types.ImplementedSporksMap[id] = true
}

func sendZts(from, to types.Address, zts types.ZenonTokenStandard, amount, fee *big.Int, data []byte) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       from,
		ToAddress:     to,
		TokenStandard: zts,
		Amount:        amount,
		Data:          data,
		Fee:           fee,
	}
}
func TestFees(t *testing.T) {
	z := mock.NewMockZenon(t)
	//ledgerApi := api.NewLedgerApi(z)
	defer z.StopPanic()
	z.InsertMomentumsTo(10)

	sendAmount := big.NewInt(5 * g.Zexp)
	fee := big.NewInt(0)

	z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 12000*g.Zexp)
	z.InsertSendBlock(sendZts(g.User1.Address, g.User2.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{}), nil, mock.SkipVmChanges)
	z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 11995*g.Zexp)
	insertMomentums(z, 2)

	// fee should not be deducted
	fee = big.NewInt(1 * g.Zexp)
	z.InsertSendBlock(sendZts(g.User1.Address, g.User2.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{}), nil, mock.SkipVmChanges)
	z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 11990*g.Zexp)
	insertMomentums(z, 2)
	autoreceive(t, z, g.User2.Address)
	activateFee(z)
	insertMomentums(z, 15)

	defer z.CallContract(&nom.AccountBlock{
		Address:       g.User1.Address,
		ToAddress:     types.TokenContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.TokenIssueAmount,
		Data: definition.ABIToken.PackMethodPanic(definition.IssueMethodName,
			"test.tok3n_na-m3",  //param.TokenName
			"TEST",              //param.TokenSymbol
			"",                  //param.TokenDomain
			big.NewInt(100000),  //param.TotalSupply
			big.NewInt(1000000), //param.MaxSupply
			uint8(1),            //param.Decimals
			true,                //param.IsMintable
			true,                //param.IsBurnable
			false,               //param.IsUtility
		),
	}).Error(t, nil)

	sendAmount = big.NewInt(10 * g.Zexp)
	fee = big.NewInt(1 * g.Zexp)
	//z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 11989*g.Zexp)
	sendBlock := sendZts(g.User1.Address, g.User2.Address, types.QsrTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)
	sendBlock = sendZts(g.User2.Address, g.User1.Address, types.QsrTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)
	sendBlock = sendZts(g.User2.Address, g.User3.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)
	z.InsertNewMomentum()

	sendAmount = big.NewInt(15 * g.Zexp)
	fee = big.NewInt(0 * g.Zexp)
	//z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 11988*g.Zexp)
	sendBlock = sendZts(g.User1.Address, g.User2.Address, types.QsrTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)

	sendBlock = sendZts(g.User2.Address, g.User3.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)

	fee = big.NewInt(1 * g.Zexp)
	sendBlock = sendZts(g.User3.Address, g.User1.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)

	sendAmount = big.NewInt(5000 * g.Zexp)
	sendBlock = sendZts(g.User1.Address, g.User3.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)
	autoreceive(t, z, g.User1.Address)
	autoreceive(t, z, g.User2.Address)
	sendBlock = sendZts(g.User2.Address, g.User3.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)

	sendBlock = sendZts(g.User1.Address, g.User3.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
	z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)

	autoreceive(t, z, g.User3.Address)
	sendAmount = big.NewInt(1 * g.Zexp / 10)
	fee = big.NewInt(1 * g.Zexp / 200)
	for i := 0; i < 750; i++ {
		sendBlock = sendZts(g.User3.Address, g.User4.Address, types.ZnnTokenStandard, sendAmount, fee, []byte{})
		z.InsertSendBlock(sendBlock, nil, mock.SkipVmChanges)
	}
	dataSendBlock, err := sendBlock.Serialize()
	common.DealWithErr(err)
	fmt.Println("dataSendBlock:", len(dataSendBlock))

	insertMomentums(z, 1)
	//z.ExpectBalance(g.User1.Address, types.ZnnTokenStandard, 11987*g.Zexp)

	ledgerApi := api.NewLedgerApi(z)
	frMom, _ := ledgerApi.GetFrontierMomentum()
	fmt.Println(len(frMom.Content))
	for _, bH := range frMom.Content {
		block, _ := ledgerApi.GetAccountBlockByHash(bH.Hash)
		fmt.Printf("Type: %d, Fee: %s, TotalPlasma: %d, Hash: %s, Address: %s\n", block.BlockType, block.Fee.String(), block.TotalPlasma, block.Hash.String(), whatAddress(block.Address))
	}
	momData, err := frMom.Serialize()
	common.DealWithErr(err)
	fmt.Println("momData:", len(momData))
}

func whatAddress(address types.Address) string {
	switch address.String() {
	case g.User1.Address.String():
		return "User1"
	case g.User2.Address.String():
		return "User2"
	case g.User3.Address.String():
		return "User3"
	case g.User4.Address.String():
		return "User4"
	case g.User5.Address.String():
		return "User5"
	default:
		return "contract"
	}
}
