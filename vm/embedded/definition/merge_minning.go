package definition

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/db"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/abi"
	"github.com/zenon-network/go-zenon/vm/constants"
	"math/big"
	"strings"
	"time"
)

const (
	jsonMergeMining = `
	[
		{"type":"function","name":"SetInitialBitcoinBlockHeader", "inputs":[
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"bits","type":"uint32"},
			{"name":"nonce","type":"uint32"},
			{"name":"height","type":"uint32"},
			{"name":"workSum","type":"uint256"}
		]},

		{"type":"function","name":"AddBitcoinBlockHeader", "inputs":[
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"bits","type":"uint32"},
			{"name":"nonce","type":"uint32"}
		]},

		{"type":"variable","name":"headerChainInfo","inputs":[
			{"name":"administrator","type":"address"},
			{"name":"tip","type":"hash"},
			{"name":"tipHeight","type":"uint32"},
			{"name":"tipWorkSum","type":"uint256"}
		]},

		{"type":"variable","name":"blockHeader","inputs":[
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"bits","type":"uint32"},
			{"name":"nonce","type":"uint32"},
			{"name":"height","type":"uint32"},
			{"name":"workSum","type":"uint256"}
		]}
	]`

	SetInitialBitcoinBlockHeaderMethodName = "SetInitialBitcoinBlockHeader"
	AddBitcoinBlockHeaderMethodName        = "AddBitcoinBlockHeader"
	//
	//SetNetworkMetadataMethodName = "SetNetworkMetadata"
	//SetBridgeMetadataMethodName  = "SetBridgeMetadata"
	//
	headerChainInfoVariableName = "headerChainInfo"
	blockHeaderVariableName     = "blockHeader"
	//networkInfoVariableName      = "networkInfo"
	//feeTokenPairVariableName     = "feeTokenPair"
	//tokenPairVariableName        = "tokenPair"
)

var (
	ABIMergeMining = abi.JSONToABIContract(strings.NewReader(jsonMergeMining))

	HeaderChainInfoPrefix = []byte{1}
	BlockHeaderKeyPrefix  = []byte{2}
)

type HeaderChainInfoVariable struct {
	// Administrator address
	Administrator types.Address `json:"administrator"`
	// Latest block of the main chain
	Tip types.Hash `json:"tip"`
	// Tip height
	TipHeight uint32 `json:"tipHeight"`
	// Most accumulated proof of work of the tip
	TipWorkSum *big.Int `json:"tipWorkSum"`
}

func (b *HeaderChainInfoVariable) Save(context db.DB) error {
	data, err := ABIMergeMining.PackVariable(
		headerChainInfoVariableName,
		b.Administrator,
		b.Tip,
		b.TipHeight,
		b.TipWorkSum,
	)
	if err != nil {
		return err
	}
	return context.Put(
		HeaderChainInfoPrefix,
		data,
	)
}
func parseHeaderChainInfoVariable(data []byte) (*HeaderChainInfoVariable, error) {
	if len(data) > 0 {
		headerChainInfo := new(HeaderChainInfoVariable)
		if err := ABIMergeMining.UnpackVariable(headerChainInfo, headerChainInfoVariableName, data); err != nil {
			return nil, err
		}
		return headerChainInfo, nil
	} else {
		return &HeaderChainInfoVariable{
			Administrator: constants.InitialMergeMiningAdministrator,
			Tip:           types.ZeroHash,
			TipHeight:     0,
			TipWorkSum:    big.NewInt(0),
		}, nil
	}
}
func GetHeaderChainInfoVariableVariable(context db.DB) (*HeaderChainInfoVariable, error) {
	if data, err := context.Get(HeaderChainInfoPrefix); err != nil {
		return nil, err
	} else {
		return parseHeaderChainInfoVariable(data)
	}
}

type BaseHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32 `json:"version"`

	// Hash of the previous block header in the block chain.
	PrevBlock types.Hash `json:"prevBlock"`

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot types.Hash `json:"merkleRoot"`

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp uint32 `json:"timestamp"`

	// Difficulty target for the block.
	Bits uint32 `json:"bits"`

	// Nonce used to generate the block.
	Nonce uint32 `json:"nonce"`
}

// BlockHash computes the block identifier hash for the given block header.
func (b *BaseHeader) BlockHash() types.Hash {
	prevBlock, err := chainhash.NewHashFromStr(b.PrevBlock.String())
	if err != nil {
		return types.ZeroHash
	}
	merkleRoot, err := chainhash.NewHashFromStr(b.MerkleRoot.String())
	if err != nil {
		return types.ZeroHash
	}

	blockHeader := &wire.BlockHeader{
		Version:    b.Version,
		PrevBlock:  *prevBlock,
		MerkleRoot: *merkleRoot,
		Timestamp:  time.Unix(int64(b.Timestamp), 0),
		Bits:       b.Bits,
		Nonce:      b.Nonce,
	}
	return types.HexToHashPanic(blockHeader.BlockHash().String())
}

type BlockHeaderVariable struct {
	BaseHeader
	Height  uint32     `json:"height"`
	WorkSum *big.Int   `json:"workSum"`
	Hash    types.Hash `json:"hash"`
}

func (b *BlockHeaderVariable) Save(context db.DB) error {
	data, err := ABIMergeMining.PackVariable(
		blockHeaderVariableName,
		b.Version,
		b.PrevBlock,
		b.MerkleRoot,
		b.Timestamp,
		b.Bits,
		b.Nonce,
		b.Height,
		b.WorkSum,
	)
	if err != nil {
		return err
	}
	return context.Put(
		GetBlockHeaderKey(b.Hash),
		data,
	)
}

func GetBlockHeaderKey(hash types.Hash) []byte {
	return common.JoinBytes(BlockHeaderKeyPrefix, hash.Bytes())
}

func parseBlockHeaderVariable(data []byte) (*BlockHeaderVariable, error) {
	if len(data) > 0 {
		blockHeader := new(BlockHeaderVariable)
		if err := ABIMergeMining.UnpackVariable(blockHeader, blockHeaderVariableName, data); err != nil {
			return nil, err
		}
		return blockHeader, nil
	} else {
		return nil, constants.ErrDataNonExistent
	}
}
func GetBlockHeaderVariable(context db.DB, hash types.Hash) (*BlockHeaderVariable, error) {
	if data, err := context.Get(GetBlockHeaderKey(hash)); err != nil {
		return nil, err
	} else {
		block, err := parseBlockHeaderVariable(data)
		if err != nil {
			return nil, err
		}
		block.Hash = hash
		return block, nil
	}
}
