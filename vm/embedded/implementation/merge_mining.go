package implementation

import (
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/vm/vm_context"
	"math/big"
	"reflect"
)

type SetInitialBitcoinBlock struct {
	MethodName string
}

func (p *SetInitialBitcoinBlock) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *SetInitialBitcoinBlock) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.BlockHeaderVariable)
	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}
	// todo validate input

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Bits, param.Nonce, param.Height, param.WorkSum)
	return err
}
func (p *SetInitialBitcoinBlock) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	headerChainInfo, err := definition.GetHeaderChainInfoVariableVariable(context.Storage())
	common.DealWithErr(err)
	if !reflect.DeepEqual(sendBlock.Address.Bytes(), headerChainInfo.Administrator.Bytes()) {
		return nil, constants.ErrPermissionDenied
	}

	param := new(definition.BlockHeaderVariable)
	err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	common.DealWithErr(err)
	blockHash := param.BaseHeader.BlockHash()
	if err := param.Hash.SetBytes(blockHash.Bytes()); err != nil {
		return nil, constants.ErrForbiddenParam
	}

	// It means merge mining has not been initialised and the administrator must set the starting block
	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		headerChainInfo.Tip = param.Hash
		headerChainInfo.TipHeight = param.Height
		headerChainInfo.TipWorkSum.Set(param.WorkSum)
		common.DealWithErr(headerChainInfo.Save(context.Storage()))
		common.DealWithErr(param.Save(context.Storage()))
	}

	return nil, nil
}

type AddBitcoinBlockHeaderMethod struct {
	MethodName string
}

func (p *AddBitcoinBlockHeaderMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *AddBitcoinBlockHeaderMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.BlockHeaderVariable)

	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}
	// todo validate input

	if block.Amount.Sign() <= 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Bits, param.Nonce, param.WorkSum)
	return err
}
func (p *AddBitcoinBlockHeaderMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}
	param := new(definition.BlockHeaderVariable)
	err := definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	common.DealWithErr(err)
	blockHash := param.BaseHeader.BlockHash()
	if err := param.Hash.SetBytes(blockHash.Bytes()); err != nil {
		return nil, constants.ErrForbiddenParam
	}

	headerChainInfo, err := definition.GetHeaderChainInfoVariableVariable(context.Storage())
	common.DealWithErr(err)

	// It means merge mining has not been initialised and the administrator must set the starting block
	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		if !reflect.DeepEqual(sendBlock.Address.Bytes(), headerChainInfo.Administrator.Bytes()) {
			return nil, constants.ErrPermissionDenied
		}
		headerChainInfo.Tip = param.Hash
		headerChainInfo.TipHeight = param.Height
		headerChainInfo.TipWorkSum.Set(param.WorkSum)
	} else {

	}

	common.DealWithErr(headerChainInfo.Save(context.Storage()))

	return nil, nil
}
