package tests

import (
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/pkg/errors"
	g "github.com/zenon-network/go-zenon/chain/genesis/mock"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api/embedded"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/zenon/mock"
	"math/big"
	"testing"
	"time"
)

func activateMergeMiningSpork(z mock.MockZenon) {
	z.InsertSendBlock(&nom.AccountBlock{
		Address:   g.Spork.Address,
		ToAddress: types.SporkContract,
		Data: definition.ABISpork.PackMethodPanic(definition.SporkCreateMethodName,
			"spork-merge-mining",              // name
			"activate spork for merge-mining", // description
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
	types.MergeMiningSpork.SporkId = id
	types.ImplementedSporksMap[id] = true

	constants.MinAdministratorDelay = 20
	constants.MinSoftDelay = 10
	constants.MinGuardians = 4
}

// Activate spork
func mergeMiningStep0(t *testing.T, z mock.MockZenon) {
	activateMergeMiningSpork(z)
	z.InsertMomentumsTo(10)

	constants.InitialMergeMiningAdministrator = g.User5.Address

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	common.Json(mergeMiningAPI.GetMergeMiningInfo()).Equals(t, `
{
	"administrator": "z1qqaswvt0e3cc5sm7lygkyza9ra63cr8e6zre09",
	"compressedTssECDSAPubKey": "",
	"decompressedTssECDSAPubKey": "",
	"metadata": ""
}`)

	common.Json(mergeMiningAPI.GetHeaderChainInfo()).Equals(t, `
{
	"tip": "0000000000000000000000000000000000000000000000000000000000000000",
	"tipHeight": 0,
	"tipWorkSum": 0
}`)
}

// Activate spork
// Sets guardians
func mergeMiningStep1(t *testing.T, z mock.MockZenon) {
	mergeMiningStep0(t, z)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	securityInfo, err := mergeMiningAPI.GetSecurityInfo()
	common.DealWithErr(err)

	guardians := []types.Address{g.User1.Address, g.User2.Address, g.User3.Address, g.User4.Address, g.User5.Address}
	nominateGuardiansMergeMining(t, z, g.User5.Address, guardians, securityInfo.AdministratorDelay)

	common.Json(mergeMiningAPI.GetSecurityInfo()).Equals(t, `
{
	"guardians": [
		"z1qqaswvt0e3cc5sm7lygkyza9ra63cr8e6zre09",
		"z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx",
		"z1qraz4ermhhua89a0h0gxxan4lnzrfutgs6xxe2",
		"z1qrs2lpccnsneglhnnfwvlsj0qncnxjnwlfmjac",
		"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"
	],
	"guardiansVotes": [
		"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
		"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
		"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
		"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
		"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f"
	],
	"administratorDelay": 20,
	"softDelay": 10
}`)
	common.Json(mergeMiningAPI.GetTimeChallengesInfoMergeMining()).Equals(t, `
{
	"count": 1,
	"list": [
		{
			"MethodName": "NominateGuardians",
			"ParamsHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"ChallengeStartHeight": 11
		}
	]
}`)
}

// Activate spork
// Sets guardians
// Sets tss ecdsa public key
func mergeMiningStep2(t *testing.T, z mock.MockZenon) {
	mergeMiningStep1(t, z)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	securityInfo, err := mergeMiningAPI.GetSecurityInfo()
	common.DealWithErr(err)
	tssPubKey := "AsAQx1M3LVXCuozDOqO5b9adj/PItYgwZFG/xTDBiZzT" // priv tuSwrTEUyJI1/3y5J8L8DSjzT/AQG2IK3JG+93qhhhI=
	changeTssMergeMining(t, z, g.User5.Address, tssPubKey, securityInfo.SoftDelay)

	common.Json(mergeMiningAPI.GetMergeMiningInfo()).Equals(t, `
{
	"administrator": "z1qqaswvt0e3cc5sm7lygkyza9ra63cr8e6zre09",
	"compressedTssECDSAPubKey": "AsAQx1M3LVXCuozDOqO5b9adj/PItYgwZFG/xTDBiZzT",
	"decompressedTssECDSAPubKey": "BMAQx1M3LVXCuozDOqO5b9adj/PItYgwZFG/xTDBiZzTnQAT1qOPAkuPzu6yoewss9XbnTmZmb9JQNGXmkPYtK4=",
	"metadata": ""
}`)
	common.Json(mergeMiningAPI.GetTimeChallengesInfoMergeMining()).Equals(t, `
{
	"count": 2,
	"list": [
		{
			"MethodName": "NominateGuardians",
			"ParamsHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"ChallengeStartHeight": 11
		},
		{
			"MethodName": "ChangeTssECDSAPubKey",
			"ParamsHash": "0000000000000000000000000000000000000000000000000000000000000000",
			"ChallengeStartHeight": 37
		}
	]
}`)
}

// Activate spork
// Sets guardians
// Sets tss ecdsa public key
// Initialize starting block
func mergeMiningStep3(t *testing.T, z mock.MockZenon) {
	mergeMiningStep2(t, z)

	blockHash := types.HexToHashPanic("00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4")
	blockHeader := definition.BlockHeaderVariable{
		BaseHeader: definition.BaseHeader{
			Version:    644612096,
			PrevBlock:  types.HexToHashPanic("0000000000000000000090937d63bfb7b27a1cf1073d6bd309195c62753c87b3"),
			MerkleRoot: types.HexToHashPanic("a76c7189bc34c9614d8848cc4074037db2f7cab329f646a23ad27a5628a573b1"),
			Timestamp:  1712573645,
			Bits:       386097875,
			Nonce:      1731415048,
		},
		Hash:    blockHash,
		Height:  838288,
		WorkSum: blockchain.CalcWork(386097875),
	}

	defer z.CallContract(setInitialBitcoinBlockHeader(g.User5.Address, blockHeader)).Error(t, nil)
	insertMomentums(z, 2)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	common.Json(mergeMiningAPI.GetHeaderChainInfo()).Equals(t, `
{
	"tip": "00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4",
	"tipHeight": 838288,
	"tipWorkSum": 357033182884110630099744
}`)

	common.Json(mergeMiningAPI.GetBlockHeader(blockHash)).Equals(t, `
{
	"version": 644612096,
	"prevBlock": "0000000000000000000090937d63bfb7b27a1cf1073d6bd309195c62753c87b3",
	"merkleRoot": "a76c7189bc34c9614d8848cc4074037db2f7cab329f646a23ad27a5628a573b1",
	"timestamp": 1712573645,
	"bits": 386097875,
	"nonce": 1731415048,
	"height": 838288,
	"workSum": 357033182884110630099744,
	"hash": "00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4"
}`)
}

// Activate spork
// Sets guardians
// Sets tss ecdsa public key
// Initialize starting block
// Set share chain info
func mergeMiningStep4(t *testing.T, z mock.MockZenon) {
	mergeMiningStep3(t, z)

	id := uint32(1)
	difficulty := big.NewInt(10000)
	rewardMultiplier := uint32(1)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	securityInfo, err := mergeMiningAPI.GetSecurityInfo()
	common.DealWithErr(err)
	setShareChain(t, z, g.User5.Address, id, difficulty, rewardMultiplier, securityInfo.SoftDelay)

	common.Json(mergeMiningAPI.GetShareChainInfo(1)).Equals(t, `
{
	"id": 1,
	"difficulty": 10000,
	"rewardMultiplier": 1
}`)
}

func mergeMiningStep5(t *testing.T, z mock.MockZenon) {
	mergeMiningStep4(t, z)

	workSum, ok := big.NewInt(0).SetString("83126997340024", 10)
	if !ok {
		common.DealWithErr(errors.New("work conversion error"))
	}

	blockHash := types.HexToHashPanic("000000000000000000000142eb893b9b422b5cf60ab352c9a8a2166f0c33c3d2")
	blockHeader := definition.BlockHeaderVariable{
		BaseHeader: definition.BaseHeader{
			Version:    551550976,
			PrevBlock:  types.HexToHashPanic("00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4"),
			MerkleRoot: types.HexToHashPanic("7fad6c9b3fd2af85f631b5c0e13d6653a9e5ff268c60a78138a98da5e4cddf69"),
			Timestamp:  1712575802,
			Bits:       386097875,
			Nonce:      2118989352,
		},
		Hash:    blockHash,
		Height:  838289,
		WorkSum: workSum,
	}
	fmt.Println(blockHeader.BlockHash().String())
	defer z.CallContract(addBitcoinBlockHeader(g.User5.Address, blockHeader)).Error(t, nil)
	insertMomentums(z, 2)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	common.Json(mergeMiningAPI.GetHeaderChainInfo()).Equals(t, `
{
	"tip": "000000000000000000000142eb893b9b422b5cf60ab352c9a8a2166f0c33c3d2",
	"tipHeight": 838289,
	"tipWorkSum": 714066365768221260199488
}`)

	common.Json(mergeMiningAPI.GetBlockHeader(blockHash)).Equals(t, `
{
	"version": 551550976,
	"prevBlock": "00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4",
	"merkleRoot": "7fad6c9b3fd2af85f631b5c0e13d6653a9e5ff268c60a78138a98da5e4cddf69",
	"timestamp": 1712575802,
	"bits": 386097875,
	"nonce": 2118989352,
	"height": 838289,
	"workSum": 714066365768221260199488,
	"hash": "000000000000000000000142eb893b9b422b5cf60ab352c9a8a2166f0c33c3d2"
}`)
}

func TestMergeMining(t *testing.T) {
	z := mock.NewMockZenonWithCustomEpochDuration(t, time.Hour)
	defer z.StopPanic()
	//defer z.SaveLogs(common.EmbeddedLogger).Equals(t, ``)

	mergeMiningStep5(t, z)
}

func setInitialBitcoinBlockHeader(from types.Address, blockHeader definition.BlockHeaderVariable) *nom.AccountBlock {
	prevBlock := types.HexToHashPanic(blockHeader.PrevBlock.String())
	merkleRoot := types.HexToHashPanic(blockHeader.MerkleRoot.String())
	return &nom.AccountBlock{
		Address:       from,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.SetInitialBitcoinBlockHeaderMethodName,
			blockHeader.Version,
			prevBlock,
			merkleRoot,
			blockHeader.Timestamp,
			blockHeader.Bits,
			blockHeader.Nonce,
			blockHeader.Height,
			big.NewInt(0).Set(blockHeader.WorkSum),
		),
	}
}

func addBitcoinBlockHeader(from types.Address, blockHeader definition.BlockHeaderVariable) *nom.AccountBlock {
	prevBlock := types.HexToHashPanic(blockHeader.PrevBlock.String())
	merkleRoot := types.HexToHashPanic(blockHeader.MerkleRoot.String())
	return &nom.AccountBlock{
		Address:       from,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.AddBitcoinBlockHeaderMethodName,
			blockHeader.Version,
			prevBlock,
			merkleRoot,
			blockHeader.Timestamp,
			blockHeader.Bits,
			blockHeader.Nonce,
			blockHeader.WorkSum,
		),
	}
}

func setShareChainStep(administrator types.Address, id uint32, difficulty *big.Int, rewardMultiplier uint32) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.SetShareChainMethodName,
			id,
			difficulty,
			rewardMultiplier,
		),
	}
}

func setShareChain(t *testing.T, z mock.MockZenon, administrator types.Address, id uint32, difficulty *big.Int, rewardMultiplier uint32, delay uint64) {
	defer z.CallContract(setShareChainStep(administrator, id, difficulty, rewardMultiplier)).Error(t, nil)
	insertMomentums(z, 2)

	frMom, err := z.Chain().GetFrontierMomentumStore().GetFrontierMomentum()
	common.DealWithErr(err)
	z.InsertMomentumsTo(frMom.Height + delay + 2)

	defer z.CallContract(setShareChainStep(administrator, id, difficulty, rewardMultiplier)).Error(t, nil)
	insertMomentums(z, 2)
}

func nominateGuardiansStepMergeMining(administrator types.Address, guardians []types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.NominateGuardiansMethodName,
			guardians),
	}
}

func nominateGuardiansMergeMining(t *testing.T, z mock.MockZenon, administrator types.Address, guardians []types.Address, delay uint64) {
	defer z.CallContract(nominateGuardiansStepMergeMining(administrator, guardians)).Error(t, nil)
	insertMomentums(z, 2)

	frMom, err := z.Chain().GetFrontierMomentumStore().GetFrontierMomentum()
	common.DealWithErr(err)
	z.InsertMomentumsTo(frMom.Height + delay + 2)

	defer z.CallContract(nominateGuardiansStepMergeMining(administrator, guardians)).Error(t, nil)
	insertMomentums(z, 2)
}

func changeTssStepMergeMining(administrator types.Address, newTssPublicKey string) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.ChangeTssECDSAPubKeyMethodName,
			newTssPublicKey, "", ""),
	}
}

func changeTssMergeMining(t *testing.T, z mock.MockZenon, administrator types.Address, newTssPublicKey string, delay uint64) {
	defer z.CallContract(changeTssStepMergeMining(administrator, newTssPublicKey)).Error(t, nil)
	insertMomentums(z, 2)

	frMom, err := z.Chain().GetFrontierMomentumStore().GetFrontierMomentum()
	common.DealWithErr(err)
	z.InsertMomentumsTo(frMom.Height + delay + 2)

	defer z.CallContract(changeTssStepMergeMining(administrator, newTssPublicKey)).Error(t, nil)
	insertMomentums(z, 2)
}

func changeAdministratorStepMergeMining(administrator types.Address, newAdministrator types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.ChangeAdministratorMethodName,
			newAdministrator),
	}
}

func changeAdministratorMergeMining(t *testing.T, z mock.MockZenon, administrator types.Address, newAdministrator types.Address, delay uint64) {
	defer z.CallContract(changeAdministratorStepMergeMining(administrator, newAdministrator)).Error(t, nil)
	insertMomentums(z, 2)

	frMom, err := z.Chain().GetFrontierMomentumStore().GetFrontierMomentum()
	common.DealWithErr(err)
	z.InsertMomentumsTo(frMom.Height + delay + 2)

	defer z.CallContract(changeAdministratorStepMergeMining(administrator, newAdministrator)).Error(t, nil)
	insertMomentums(z, 2)
}

func activateEmergencyMergeMining(administrator types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data:          definition.ABIMergeMining.PackMethodPanic(definition.EmergencyMethodName),
	}
}

func setHeaderChainMetadata(administrator types.Address, metadata string) *nom.AccountBlock {
	return &nom.AccountBlock{
		Address:       administrator,
		ToAddress:     types.MergeMiningContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(0),
		Data: definition.ABIMergeMining.PackMethodPanic(definition.SetBridgeMetadataMethodName,
			metadata),
	}
}
