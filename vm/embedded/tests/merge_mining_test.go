package tests

import (
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
}

// Activate spork
func mergeMiningStep0(t *testing.T, z mock.MockZenon) {
	activateMergeMiningSpork(z)
	z.InsertMomentumsTo(10)

	constants.InitialMergeMiningAdministrator = g.User5.Address

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	common.Json(mergeMiningAPI.GetHeaderChainInfo()).Equals(t, `
{
	"administrator": "z1qqaswvt0e3cc5sm7lygkyza9ra63cr8e6zre09",
	"tip": "0000000000000000000000000000000000000000000000000000000000000000",
	"tipHeight": 0,
	"tipWorkSum": 0
}`)
}

// Initialize starting block
func mergeMiningStep1(t *testing.T, z mock.MockZenon) {
	mergeMiningStep0(t, z)

	//prevBlockHash, _ := chainhash.NewHashFromStr("0000000000000000000090937d63bfb7b27a1cf1073d6bd309195c62753c87b3")
	//merkleRoot, _ := chainhash.NewHashFromStr("a76c7189bc34c9614d8848cc4074037db2f7cab329f646a23ad27a5628a573b1")
	//bh := wire.BlockHeader{
	//	Version:    644612096,
	//	PrevBlock:  *prevBlockHash,
	//	MerkleRoot: *merkleRoot,
	//	Timestamp:  time.Unix(1712573645, 0),
	//	Bits:       386097875,
	//	Nonce:      1731415048,
	//}
	//fmt.Println("Computed from wire: ", bh.BlockHash().String())
	//
	//bhh := definition.BaseHeader{
	//	Version:    644612096,
	//	PrevBlock:  types.HexToHashPanic("0000000000000000000090937d63bfb7b27a1cf1073d6bd309195c62753c87b3"),
	//	MerkleRoot: types.HexToHashPanic("a76c7189bc34c9614d8848cc4074037db2f7cab329f646a23ad27a5628a573b1"),
	//	Timestamp:  1712573645,
	//	Bits:       386097875,
	//	Nonce:      1731415048,
	//}
	//fmt.Println("Computed from base header: ", bhh.BlockHash().String())
	//return
	workSum, ok := big.NewInt(0).SetString("83126997340024", 10)
	if !ok {
		common.DealWithErr(errors.New("work conversion error"))
	}

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
		WorkSum: workSum,
	}

	defer z.CallContract(setInitialBitcoinBlockHeader(g.User5.Address, blockHeader)).Error(t, nil)
	insertMomentums(z, 2)

	mergeMiningAPI := embedded.NewMergeMiningApi(z)
	common.Json(mergeMiningAPI.GetHeaderChainInfo()).Equals(t, `
{
	"administrator": "z1qqaswvt0e3cc5sm7lygkyza9ra63cr8e6zre09",
	"tip": "00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4",
	"tipHeight": 838288,
	"tipWorkSum": 83126997340024
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
	"workSum": 83126997340024,
	"hash": "00000000000000000001052825fecaf9987861781cb11af3639603a381db34e4"
}`)
}

func TestMergeMining(t *testing.T) {
	z := mock.NewMockZenonWithCustomEpochDuration(t, time.Hour)
	defer z.StopPanic()
	//defer z.SaveLogs(common.EmbeddedLogger).Equals(t, ``)

	mergeMiningStep1(t, z)
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
		),
	}
}
