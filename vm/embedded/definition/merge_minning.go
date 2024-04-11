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
			{"name":"height","type":"uint32"}
		]},

		{"type":"function","name":"SetShareChain","inputs":[
			{"name":"id","type":"uint8"},
			{"name":"bits","type":"uint32"},
			{"name":"rewardMultiplier","type":"uint32"}
		]},

		{"type":"function","name":"AddBitcoinBlockHeader", "inputs":[
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"bits","type":"uint32"},
			{"name":"nonce","type":"uint32"}
		]},

		{"type":"function","name":"AddShare", "inputs":[
			{"name":"shareChainId","type":"uint8"},
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"nonce","type":"uint32"},
			{"name":"proof","type":"hash"},
			{"name":"prooff","type":"hash"},
			{"name":"proofff","type":"hash"},
			{"name":"prooffff","type":"hash"},
			{"name":"additionalData","type":"hash"}
		]},

		{"type":"function","name":"Emergency","inputs":[]},
		{"type":"function","name":"NominateGuardians","inputs":[
			{"name":"guardians","type":"address[]"}
		]},
		{"type":"function","name":"ChangeAdministrator","inputs":[
			{"name":"administrator","type":"address"}
		]},
		{"type":"function","name":"ProposeAdministrator","inputs":[
			{"name":"address","type":"address"}
		]},
		{"type":"function","name":"ChangeTssECDSAPubKey","inputs":[
			{"name":"pubKey","type":"string"},
			{"name":"oldPubKeySignature","type":"string"},
			{"name":"newPubKeySignature","type":"string"}
		]},
		{"type":"function","name":"SetMergeMiningMetadata","inputs":[
			{"name":"metadata","type":"string"}
		]},

		{"type":"variable","name":"mergeMiningInfo","inputs":[
			{"name":"administrator","type":"address"},
			{"name":"compressedTssECDSAPubKey","type":"string"},
			{"name":"decompressedTssECDSAPubKey","type":"string"},
			{"name":"metadata","type":"string"}
		]},
		
		{"type":"variable","name":"headerChainInfo","inputs":[
			{"name":"tip","type":"hash"},
			{"name":"tipHeight","type":"uint32"},
			{"name":"tipWorkSum","type":"uint256"}
		]},

		{"type":"variable","name":"shareChainInfo","inputs":[
			{"name":"bits","type":"uint32"},
			{"name":"rewardMultiplier","type":"uint32"}
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
		]},

		{"type":"variable","name":"share","inputs":[
			{"name":"shareChainId","type":"uint8"},
			{"name":"witness","type":"bool"},
			{"name":"version","type":"int32"},
			{"name":"prevBlock","type":"hash"},
			{"name":"merkleRoot","type":"hash"},
			{"name":"timestamp","type":"uint32"},
			{"name":"nonce","type":"uint32"},
			{"name":"proof","type":"hash"},
			{"name":"prooff","type":"hash"},
			{"name":"proofff","type":"hash"},
			{"name":"prooffff","type":"hash"},
			{"name":"additionalData","type":"hash"}
		]}
	]`

	SetInitialBitcoinBlockHeaderMethodName = "SetInitialBitcoinBlockHeader"
	AddBitcoinBlockHeaderMethodName        = "AddBitcoinBlockHeader"
	SetMergeMiningMetadataMethodName       = "SetMergeMiningMetadata"
	SetShareChainMethodName                = "SetShareChain"
	AddShareMethodName                     = "AddShare"

	mergeMiningInfoVariableName = "mergeMiningInfo"
	headerChainInfoVariableName = "headerChainInfo"
	shareChainInfoVariableName  = "shareChainInfo"
	blockHeaderVariableName     = "blockHeader"
	shareVariableName           = "share"
)

var (
	ABIMergeMining = abi.JSONToABIContract(strings.NewReader(jsonMergeMining))

	MergeMiningInfoPrefix = []byte{1}
	HeaderChainInfoPrefix = []byte{2}
	ShareChainInfoPrefix  = []byte{3}
	ShareKeyPrefix        = []byte{4}
	BlockHeaderKeyPrefix  = []byte{5}
)

type Share struct {
	ShareChainId   uint8      `json:"shareChainId"`
	Witness        bool       `json:"witness"`
	Version        int32      `json:"version"`
	PrevBlock      types.Hash `json:"prevBlock"`
	MerkleRoot     types.Hash `json:"merkleRoot"`
	Timestamp      uint32     `json:"timestamp"`
	Nonce          uint32     `json:"nonce"`
	Proof          types.Hash `json:"proof"`
	Prooff         types.Hash `json:"prooff"`
	Proofff        types.Hash `json:"proofff"`
	Prooffff       types.Hash `json:"prooffff"`
	AdditionalData types.Hash `json:"additionalData"`
}

type MergeMiningInfoVariable struct {
	// Administrator address
	Administrator types.Address `json:"administrator"`
	// ECDSA pub key generated by the orchestrator from key gen ceremony
	CompressedTssECDSAPubKey   string `json:"compressedTssECDSAPubKey"`
	DecompressedTssECDSAPubKey string `json:"decompressedTssECDSAPubKey"`

	// Additional metadata
	Metadata string `json:"metadata"`
}

func (b *MergeMiningInfoVariable) Save(context db.DB) error {
	data, err := ABIMergeMining.PackVariable(
		mergeMiningInfoVariableName,
		b.Administrator,
		b.CompressedTssECDSAPubKey,
		b.DecompressedTssECDSAPubKey,
		b.Metadata,
	)
	if err != nil {
		return err
	}
	return context.Put(
		MergeMiningInfoPrefix,
		data,
	)
}
func parseMergeMiningInfoVariable(data []byte) (*MergeMiningInfoVariable, error) {
	if len(data) > 0 {
		mergeMiningInfo := new(MergeMiningInfoVariable)
		if err := ABIMergeMining.UnpackVariable(mergeMiningInfo, mergeMiningInfoVariableName, data); err != nil {
			return nil, err
		}
		return mergeMiningInfo, nil
	} else {
		return &MergeMiningInfoVariable{
			Administrator:              constants.InitialMergeMiningAdministrator,
			CompressedTssECDSAPubKey:   "",
			DecompressedTssECDSAPubKey: "",
			Metadata:                   "",
		}, nil
	}
}
func GetMergeMiningInfoVariableVariable(context db.DB) (*MergeMiningInfoVariable, error) {
	if data, err := context.Get(MergeMiningInfoPrefix); err != nil {
		return nil, err
	} else {
		return parseMergeMiningInfoVariable(data)
	}
}

type ShareChainInfoVariable struct {
	Id               uint8  `json:"id"`
	Bits             uint32 `json:"bits"`
	RewardMultiplier uint32 `json:"rewardMultiplier"`
}

func GetShareChainKey(id uint8) []byte {
	idBytes := make([]byte, 4)
	idBytes[3] = id
	return common.JoinBytes(ShareChainInfoPrefix, idBytes)
}

func (s *ShareChainInfoVariable) Save(context db.DB) error {
	data, err := ABIMergeMining.PackVariable(
		shareChainInfoVariableName,
		s.Bits,
		s.RewardMultiplier,
	)
	if err != nil {
		return err
	}
	return context.Put(
		GetShareChainKey(s.Id),
		data,
	)
}
func parseShareChainInfoVariable(data []byte) (*ShareChainInfoVariable, error) {
	if len(data) > 0 {
		shareChainInfo := new(ShareChainInfoVariable)
		if err := ABIMergeMining.UnpackVariable(shareChainInfo, shareChainInfoVariableName, data); err != nil {
			return nil, err
		}
		return shareChainInfo, nil
	} else {
		return nil, constants.ErrDataNonExistent
	}
}
func GetShareChainInfoVariableVariable(context db.DB, id uint8) (*ShareChainInfoVariable, error) {
	if data, err := context.Get(GetShareChainKey(id)); err != nil {
		return nil, err
	} else {
		shareChainInfo, err := parseShareChainInfoVariable(data)
		if err != nil {
			return nil, err
		}
		shareChainInfo.Id = id
		return shareChainInfo, nil
	}
}

type HeaderChainInfoVariable struct {
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
			Tip:        types.ZeroHash,
			TipHeight:  0,
			TipWorkSum: big.NewInt(0),
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

// BlockHash computes the block identifier hash for the given block header.
func (b *BaseHeader) BlockHashChain() chainhash.Hash {
	prevBlock, err := chainhash.NewHashFromStr(b.PrevBlock.String())
	if err != nil {
		return chainhash.Hash{}
	}
	merkleRoot, err := chainhash.NewHashFromStr(b.MerkleRoot.String())
	if err != nil {
		return chainhash.Hash{}
	}

	blockHeader := &wire.BlockHeader{
		Version:    b.Version,
		PrevBlock:  *prevBlock,
		MerkleRoot: *merkleRoot,
		Timestamp:  time.Unix(int64(b.Timestamp), 0),
		Bits:       b.Bits,
		Nonce:      b.Nonce,
	}
	return blockHeader.BlockHash()
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
