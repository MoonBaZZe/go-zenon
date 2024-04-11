package implementation

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/crypto"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/vm/vm_context"
	"math/big"
	"reflect"
	"sort"
	"time"
)

func CheckMergeMiningInitialized(context vm_context.AccountVmContext) (*definition.MergeMiningInfoVariable, error) {
	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	common.DealWithErr(err)

	if len(mergeMiningInfo.CompressedTssECDSAPubKey) == 0 {
		return nil, constants.ErrMergeMiningNotInitialized
	}

	return mergeMiningInfo, nil
}

func CheckHeaderChainInitialized(context vm_context.AccountVmContext) (*definition.HeaderChainInfoVariable, error) {
	headerChainInfo, err := definition.GetHeaderChainInfoVariableVariable(context.Storage())
	common.DealWithErr(err)

	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		return nil, constants.ErrHeaderChainNotInitialized
	}

	return headerChainInfo, nil
}

func CanPerformActionMergeMining(context vm_context.AccountVmContext) (*definition.MergeMiningInfoVariable, *definition.HeaderChainInfoVariable, error) {
	if _, errSec := CheckSecurityInitialized(context); errSec != nil {
		return nil, nil, errSec
	} else {
		if mergeMiningInfo, errMergeMining := CheckMergeMiningInitialized(context); errMergeMining != nil {
			return nil, nil, errMergeMining
		} else {
			if headerChainInfo, errHeaderChain := CheckHeaderChainInitialized(context); errHeaderChain != nil {
				return nil, nil, errHeaderChain
			} else {
				return mergeMiningInfo, headerChainInfo, nil
			}
		}
	}
}

// We need guardians and pubKey set
type SetInitialBitcoinBlockMethod struct {
	MethodName string
}

func (p *SetInitialBitcoinBlockMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *SetInitialBitcoinBlockMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.BlockHeaderVariable)
	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}
	if errPoW := CheckProofOfWork(*param, param.Bits); errPoW != nil {
		return errPoW
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Bits, param.Nonce, param.Height)
	return err
}
func (p *SetInitialBitcoinBlockMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	_, errSec := CheckSecurityInitialized(context)
	if errSec != nil {
		return nil, errSec
	}

	mergeMiningInfo, err := CheckMergeMiningInitialized(context)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(sendBlock.Address.Bytes(), mergeMiningInfo.Administrator.Bytes()) {
		return nil, constants.ErrPermissionDenied
	}

	param := new(definition.BlockHeaderVariable)
	err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	common.DealWithErr(err)
	blockHash := param.BaseHeader.BlockHash()
	if err := param.Hash.SetBytes(blockHash.Bytes()); err != nil {
		return nil, constants.ErrForbiddenParam
	}

	headerChainInfo, err := definition.GetHeaderChainInfoVariableVariable(context.Storage())
	common.DealWithErr(err)
	// It means merge mining has not been initialised and the administrator must set the starting block
	workSum := big.NewInt(0).Set(blockchain.CalcWork(param.Bits))
	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		headerChainInfo.Tip = param.Hash
		headerChainInfo.TipHeight = param.Height
		headerChainInfo.TipWorkSum.Set(workSum)
		common.DealWithErr(headerChainInfo.Save(context.Storage()))
		param.WorkSum = big.NewInt(0).Set(workSum)
		common.DealWithErr(param.Save(context.Storage()))
	}

	return nil, nil
}

type SetShareChainMethod struct {
	MethodName string
}

func (p *SetShareChainMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *SetShareChainMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.ShareChainInfoVariable)
	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}
	// todo validate input

	if block.Amount.Sign() > 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Id, param.Bits, param.RewardMultiplier)
	return err
}
func (p *SetShareChainMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	mergeMiningInfo, _, err := CanPerformActionMergeMining(context)
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	param := new(definition.ShareChainInfoVariable)
	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data); err != nil {
		return nil, constants.ErrUnpackError
	}

	securityInfo, err := definition.GetSecurityInfoVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	paramsBytes := make([]byte, 1+4+4)
	idBytes := make([]byte, 4)
	paramsBytes = append(paramsBytes, param.Id)
	idBytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(idBytes, param.Bits)
	paramsBytes = append(paramsBytes, ecommon.LeftPadBytes(idBytes, 4)...)
	idBytes = make([]byte, 4)
	binary.LittleEndian.PutUint32(idBytes, param.RewardMultiplier)
	paramsBytes = append(paramsBytes, ecommon.LeftPadBytes(idBytes, 4)...)
	idBytes = make([]byte, 0)

	paramsHash := crypto.Hash(paramsBytes)
	if timeChallengeInfo, errTimeChallenge := TimeChallenge(context, p.MethodName, paramsHash, securityInfo.SoftDelay); errTimeChallenge != nil {
		return nil, errTimeChallenge
	} else {
		// if paramsHash is not zero it means we had a new challenge and we can't go further to save the change into local db
		if !timeChallengeInfo.ParamsHash.IsZero() {
			return nil, nil
		}
	}

	// We just save the new share chain, same as editing it
	common.DealWithErr(param.Save(context.Storage()))
	return nil, nil
}

func CheckProofOfWork(header definition.BlockHeaderVariable, targetBits uint32) error {
	hash := header.BlockHashChain()
	fmt.Println(hash.String())
	targetDifficulty := blockchain.CompactToBig(targetBits)
	fmt.Println(targetDifficulty.String())
	if targetDifficulty.Sign() <= 0 {
		return constants.ErrTargetDifficultyLessThanZero
	}
	fmt.Println(blockchain.HashToBig(&hash).String())
	if blockchain.HashToBig(&hash).Cmp(targetDifficulty) > 0 {
		return constants.ErrInvalidNonce
	}
	// The target difficulty must be less than the maximum allowed.
	if targetDifficulty.Cmp(constants.MainPowLimit) > 0 {
		return constants.ErrPowLimitExceeded
	}
	return nil
}

// CalcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func CalcEasiestDifficulty(bits uint32, duration time.Duration) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration / time.Second)
	adjustmentFactor := big.NewInt(constants.RetargetAdjustmentFactor)

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := blockchain.CompactToBig(bits)
	for durationVal > 0 && newTarget.Cmp(constants.MainPowLimit) < 0 {
		newTarget.Mul(newTarget, adjustmentFactor)
		durationVal -= constants.MaxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(constants.MainPowLimit) > 0 {
		newTarget.Set(constants.MainPowLimit)
	}

	return blockchain.BigToCompact(newTarget)
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

	if errPoW := CheckProofOfWork(*param, param.Bits); errPoW != nil {
		return errPoW
	}

	if block.Amount.Sign() > 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Bits, param.Nonce)
	return err
}
func (p *AddBitcoinBlockHeaderMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	_, headerChainInfo, err := CanPerformActionMergeMining(context)
	if err != nil {
		return nil, err
	}

	param := new(definition.BlockHeaderVariable)
	err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	common.DealWithErr(err)

	prevBlock, err := definition.GetBlockHeaderVariable(context.Storage(), param.PrevBlock)
	if err != nil {
		if !errors.Is(err, constants.ErrDataNonExistent) {
			common.DealWithErr(err)
		} else {
			return nil, constants.ErrPrevBlockNonExistent
		}
	}

	// ! If the timestamp is in the far future, one could get away with a low difficulty and spam blocks

	// todo better validate timestamp?
	if param.Timestamp < prevBlock.Timestamp {
		return nil, constants.ErrForbiddenParam
	}
	// todo is this correct?
	easiestDifficulty := CalcEasiestDifficulty(prevBlock.Bits, time.Duration(int64(param.Timestamp-prevBlock.Timestamp)))
	if easiestDifficulty > param.Bits {
		return nil, constants.ErrDifficultyLessThanMin
	}

	hash := param.BlockHash()
	if err := param.Hash.SetBytes(hash.Bytes()); err != nil {
		return nil, constants.ErrForbiddenParam
	}

	// We do not allow duplicate blocks
	if _, err = definition.GetBlockHeaderVariable(context.Storage(), hash); err != nil {
		if !errors.Is(err, constants.ErrDataNonExistent) {
			common.DealWithErr(err)
		}
	}

	workSum := big.NewInt(0).Add(prevBlock.WorkSum, blockchain.CalcWork(param.Bits))
	// We found a block with more accumulated pow then the previous max
	if workSum.Cmp(headerChainInfo.TipWorkSum) > 0 {
		headerChainInfo.Tip = hash
		headerChainInfo.TipHeight = prevBlock.Height + 1
		headerChainInfo.TipWorkSum.Set(workSum)
		common.DealWithErr(headerChainInfo.Save(context.Storage()))
	}
	param.WorkSum = big.NewInt(0).Set(workSum)
	param.Height = prevBlock.Height + 1
	common.DealWithErr(param.Save(context.Storage()))

	common.DealWithErr(headerChainInfo.Save(context.Storage()))
	return nil, nil
}

type AddShareMethod struct {
	MethodName string
}

func (p *AddShareMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *AddShareMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.Share)

	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() > 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.ShareChainId, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Nonce,
		param.Proof, param.Prooff, param.Proofff, param.Prooffff, param.AdditionalData)
	return err
}
func (p *AddShareMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	_, headerChainInfo, err := CanPerformActionMergeMining(context)
	if err != nil {
		return nil, err
	}

	param := new(definition.Share)
	err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	common.DealWithErr(err)

	shareChainInfo, err := definition.GetShareChainInfoVariableVariable(context.Storage(), param.ShareChainId)
	if err != nil {
		if !errors.Is(err, constants.ErrDataNonExistent) {
			common.DealWithErr(err)
		} else {
			return nil, constants.ErrShareChainNonExistent
		}
	}

	prevBlock, err := definition.GetBlockHeaderVariable(context.Storage(), param.PrevBlock)
	if err != nil {
		if !errors.Is(err, constants.ErrDataNonExistent) {
			common.DealWithErr(err)
		} else {
			return nil, constants.ErrPrevBlockNonExistent
		}
	}
	// Share is too far in the past
	if prevBlock.Height+1 < headerChainInfo.TipHeight {
		return nil, constants.ErrForbiddenParam
	}

	// ! If the timestamp is in the far future, one could get away with a low difficulty and spam blocks
	// todo better validate timestamp?
	if param.Timestamp < prevBlock.Timestamp {
		return nil, constants.ErrForbiddenParam
	}

	shareChainHeader := definition.BlockHeaderVariable{
		BaseHeader: definition.BaseHeader{
			Version:    param.Version,
			PrevBlock:  param.PrevBlock,
			MerkleRoot: param.MerkleRoot,
			Timestamp:  param.Timestamp,
			Bits:       prevBlock.Bits,
			Nonce:      param.Nonce,
		},
		Height:  0,
		WorkSum: nil,
		Hash:    types.Hash{},
	}

	if err := CheckProofOfWork(shareChainHeader, shareChainInfo.Bits); err != nil {
		return nil, err
	}

	// Add pow to this address
	return nil, nil
}

type NominateGuardiansMergeMiningMethod struct {
	MethodName string
}

func (p *NominateGuardiansMergeMiningMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *NominateGuardiansMergeMiningMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error

	guardians := new([]types.Address)
	if err := definition.ABIMergeMining.UnpackMethod(guardians, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	if len(*guardians) < constants.MinGuardians {
		return constants.ErrInvalidGuardians
	}
	for _, address := range *guardians {
		// we also check with this method because in the abi the checksum is not verified
		parsedAddress, err := types.ParseAddress(address.String())
		if err != nil {
			return err
		} else if parsedAddress.IsZero() {
			return constants.ErrForbiddenParam
		}
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, guardians)
	return err
}
func (p *NominateGuardiansMergeMiningMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	guardians := new([]types.Address)
	err := definition.ABIMergeMining.UnpackMethod(guardians, p.MethodName, sendBlock.Data)
	if err != nil {
		return nil, err
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	securityInfo, err := definition.GetSecurityInfoVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	sort.SliceStable(*guardians, func(i, j int) bool {
		return (*guardians)[i].String() < (*guardians)[j].String()
	})

	guardiansBytes := make([]byte, 0)
	for _, g := range *guardians {
		guardiansBytes = append(guardiansBytes, g.Bytes()...)
	}
	paramsHash := crypto.Hash(guardiansBytes)
	if timeChallengeInfo, errTimeChallenge := TimeChallenge(context, p.MethodName, paramsHash, securityInfo.AdministratorDelay); errTimeChallenge != nil {
		return nil, errTimeChallenge
	} else {
		// if paramsHash is not zero it means we had a new challenge and we can't go further to save the change into local db
		if !timeChallengeInfo.ParamsHash.IsZero() {
			return nil, nil
		}
	}

	securityInfo.Guardians = make([]types.Address, 0)
	securityInfo.GuardiansVotes = make([]types.Address, 0)
	for _, guardian := range *guardians {
		securityInfo.Guardians = append(securityInfo.Guardians, guardian)
		// append empty vote
		securityInfo.GuardiansVotes = append(securityInfo.GuardiansVotes, types.Address{})
	}

	common.DealWithErr(securityInfo.Save(context.Storage()))
	return nil, nil
}

type ProposeAdministratorMergeMiningMethod struct {
	MethodName string
}

func (p *ProposeAdministratorMergeMiningMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *ProposeAdministratorMergeMiningMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error

	address := new(types.Address)
	if err := definition.ABIMergeMining.UnpackMethod(address, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	// we also check with this method because in the abi the checksum is not verified
	parsedAddress, err := types.ParseAddress(address.String())
	if err != nil {
		return err
	} else if parsedAddress.IsZero() {
		return constants.ErrForbiddenParam
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, *address)
	return err
}
func (p *ProposeAdministratorMergeMiningMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	proposedAddress := new(types.Address)
	if err := definition.ABIMergeMining.UnpackMethod(proposedAddress, p.MethodName, sendBlock.Data); err != nil {
		return nil, constants.ErrUnpackError
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if !mergeMiningInfo.Administrator.IsZero() {
		return nil, constants.ErrNotEmergency
	}

	securityInfo, err := definition.GetSecurityInfoVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	found := false
	for idx, guardian := range securityInfo.Guardians {
		if bytes.Equal(guardian.Bytes(), sendBlock.Address.Bytes()) {
			found = true
			if err := securityInfo.GuardiansVotes[idx].SetBytes(proposedAddress.Bytes()); err != nil {
				return nil, err
			}
			break
		}
	}
	if !found {
		return nil, constants.ErrNotGuardian
	}

	votes := make(map[string]uint8)

	threshold := uint8(len(securityInfo.Guardians) / 2)
	for _, vote := range securityInfo.GuardiansVotes {
		if !vote.IsZero() {
			votes[vote.String()] += 1
			// we got a majority, so we change the administrator pub key
			if votes[vote.String()] > threshold {
				votedAddress, errParse := types.ParseAddress(vote.String())
				if errParse != nil {
					return nil, errParse
				} else if votedAddress.IsZero() {
					return nil, constants.ErrForbiddenParam
				}
				if errSet := mergeMiningInfo.Administrator.SetBytes(votedAddress.Bytes()); errSet != nil {
					return nil, errSet
				}
				common.DealWithErr(mergeMiningInfo.Save(context.Storage()))
				for idx, _ := range securityInfo.GuardiansVotes {
					securityInfo.GuardiansVotes[idx] = types.Address{}
				}
				break
			}
		}
	}
	common.DealWithErr(securityInfo.Save(context.Storage()))
	return nil, nil
}

type EmergencyMergeMiningMethod struct {
	MethodName string
}

func (p *EmergencyMergeMiningMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *EmergencyMergeMiningMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	if err := definition.ABIMergeMining.UnpackEmptyMethod(p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName)
	return err
}
func (p *EmergencyMergeMiningMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	err := definition.ABIMergeMining.UnpackEmptyMethod(p.MethodName, sendBlock.Data)
	if err != nil {
		return nil, err
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if _, err := CheckSecurityInitialized(context); err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	if errSet := mergeMiningInfo.Administrator.SetBytes(types.ZeroAddress.Bytes()); errSet != nil {
		return nil, errSet
	}
	common.DealWithErr(mergeMiningInfo.Save(context.Storage()))
	return nil, nil
}

type ChangeAdministratorMergeMiningMethod struct {
	MethodName string
}

func (p *ChangeAdministratorMergeMiningMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *ChangeAdministratorMergeMiningMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	address := new(types.Address)
	if err = definition.ABIMergeMining.UnpackMethod(address, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	// we also check with this method because in the abi the checksum is not verified
	parsedAddress, err := types.ParseAddress(address.String())
	if err != nil {
		return err
	} else if parsedAddress.IsZero() {
		return constants.ErrForbiddenParam
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, address)
	return err
}
func (p *ChangeAdministratorMergeMiningMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	address := new(types.Address)
	err := definition.ABIMergeMining.UnpackMethod(address, p.MethodName, sendBlock.Data)
	if err != nil {
		return nil, err
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	securityInfo, err := CheckSecurityInitialized(context)
	if err != nil {
		return nil, err
	}

	paramsHash := crypto.Hash(address.Bytes())
	if timeChallengeInfo, errTimeChallenge := TimeChallenge(context, p.MethodName, paramsHash, securityInfo.AdministratorDelay); errTimeChallenge != nil {
		return nil, errTimeChallenge
	} else {
		// if paramsHash is not zero it means we had a new challenge and we can't go further to save the change into local db
		if !timeChallengeInfo.ParamsHash.IsZero() {
			return nil, nil
		}
	}

	if errSet := mergeMiningInfo.Administrator.SetBytes(address.Bytes()); errSet != nil {
		return nil, err
	}
	common.DealWithErr(mergeMiningInfo.Save(context.Storage()))
	return nil, nil
}

type ChangeTssECDSAPubKeyMergeMiningMethod struct {
	MethodName string
}

func (p *ChangeTssECDSAPubKeyMergeMiningMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *ChangeTssECDSAPubKeyMergeMiningMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error
	param := new(definition.ChangeECDSAPubKeyParam)
	if err = definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}
	pubKey, err := base64.StdEncoding.DecodeString(param.PubKey)
	if err != nil {
		return constants.ErrInvalidB64Decode
	}
	if len(pubKey) != constants.CompressedECDSAPubKeyLength {
		return constants.ErrInvalidCompressedECDSAPubKeyLength
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.PubKey, param.OldPubKeySignature, param.NewPubKeySignature)
	return err
}
func (p *ChangeTssECDSAPubKeyMergeMiningMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	param := new(definition.ChangeECDSAPubKeyParam)
	err := definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	if err != nil {
		return nil, err
	}

	if _, err := CheckSecurityInitialized(context); err != nil {
		return nil, err
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	// we check it in the send block
	pubKey, _ := base64.StdEncoding.DecodeString(param.PubKey)

	X, Y := secp256k1.DecompressPubkey(pubKey)
	dPubKeyBytes := make([]byte, 1)
	dPubKeyBytes[0] = 4
	dPubKeyBytes = append(dPubKeyBytes, X.Bytes()...)
	dPubKeyBytes = append(dPubKeyBytes, Y.Bytes()...)
	newDecompressedPubKey := base64.StdEncoding.EncodeToString(dPubKeyBytes)

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	} else {
		securityInfo, err := definition.GetSecurityInfoVariable(context.Storage())
		if err != nil {
			return nil, err
		}
		paramsHash := crypto.Hash(dPubKeyBytes)
		if timeChallengeInfo, errTimeChallenge := TimeChallenge(context, p.MethodName, paramsHash, securityInfo.SoftDelay); errTimeChallenge != nil {
			return nil, errTimeChallenge
		} else {
			// if paramsHash is not zero it means we had a new challenge and we can't go further to save the change into local db
			if !timeChallengeInfo.ParamsHash.IsZero() {
				return nil, nil
			}
		}
	}

	mergeMiningInfo.CompressedTssECDSAPubKey = param.PubKey
	mergeMiningInfo.DecompressedTssECDSAPubKey = newDecompressedPubKey

	common.DealWithErr(mergeMiningInfo.Save(context.Storage()))
	return nil, nil
}

type SetMergeMiningMetadataMethod struct {
	MethodName string
}

func (p *SetMergeMiningMetadataMethod) GetPlasma(plasmaTable *constants.PlasmaTable) (uint64, error) {
	return plasmaTable.EmbeddedSimple, nil
}
func (p *SetMergeMiningMetadataMethod) ValidateSendBlock(block *nom.AccountBlock) error {
	var err error

	param := new(string)
	if err := definition.ABIMergeMining.UnpackMethod(param, p.MethodName, block.Data); err != nil {
		return constants.ErrUnpackError
	}

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	if !IsJSON(*param) {
		return constants.ErrInvalidJsonContent
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param)
	return err
}
func (p *SetMergeMiningMetadataMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
	if err := p.ValidateSendBlock(sendBlock); err != nil {
		return nil, err
	}

	param := new(string)
	err := definition.ABIMergeMining.UnpackMethod(param, p.MethodName, sendBlock.Data)
	if err != nil {
		return nil, err
	}

	mergeMiningInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != mergeMiningInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	mergeMiningInfo.Metadata = *param
	common.DealWithErr(mergeMiningInfo.Save(context.Storage()))
	return nil, nil
}
