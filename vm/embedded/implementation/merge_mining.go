package implementation

import (
	"bytes"
	"encoding/base64"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
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
	// todo validate input

	if block.Amount.Sign() != 0 {
		return constants.ErrInvalidTokenOrAmount
	}

	block.Data, err = definition.ABIMergeMining.PackMethod(p.MethodName, param.Version, param.PrevBlock, param.MerkleRoot, param.Timestamp, param.Bits, param.Nonce, param.Height, param.WorkSum)
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
	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		headerChainInfo.Tip = param.Hash
		headerChainInfo.TipHeight = param.Height
		headerChainInfo.TipWorkSum.Set(param.WorkSum)
		common.DealWithErr(headerChainInfo.Save(context.Storage()))
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
func (p *SetShareChainMethod) ReceiveBlock(context vm_context.AccountVmContext, sendBlock *nom.AccountBlock) ([]*nom.AccountBlock, error) {
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
	blockHash := param.BaseHeader.BlockHash()
	if err := param.Hash.SetBytes(blockHash.Bytes()); err != nil {
		return nil, constants.ErrForbiddenParam
	}

	// It means merge mining has not been initialised and the administrator must set the starting block
	if reflect.DeepEqual(headerChainInfo.Tip.Bytes(), types.ZeroHash) || headerChainInfo.TipHeight == 0 || headerChainInfo.TipWorkSum.Cmp(big.NewInt(0)) == 0 {
		headerChainInfo.Tip = param.Hash
		headerChainInfo.TipHeight = param.Height
		headerChainInfo.TipWorkSum.Set(param.WorkSum)
	} else {

	}

	common.DealWithErr(headerChainInfo.Save(context.Storage()))

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

	_, headerChainInfo, err := CanPerformActionMergeMining(context)
	if err != nil {
		return nil, err
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
	} else {

	}

	common.DealWithErr(headerChainInfo.Save(context.Storage()))

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

	headerChainInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != headerChainInfo.Administrator.String() {
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

	headerChainInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if !headerChainInfo.Administrator.IsZero() {
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
				if errSet := headerChainInfo.Administrator.SetBytes(votedAddress.Bytes()); errSet != nil {
					return nil, errSet
				}
				common.DealWithErr(headerChainInfo.Save(context.Storage()))
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

	headerChainInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if _, err := CheckSecurityInitialized(context); err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != headerChainInfo.Administrator.String() {
		return nil, constants.ErrPermissionDenied
	}

	if errSet := headerChainInfo.Administrator.SetBytes(types.ZeroAddress.Bytes()); errSet != nil {
		return nil, errSet
	}
	common.DealWithErr(headerChainInfo.Save(context.Storage()))
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

	headerChainInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	if sendBlock.Address.String() != headerChainInfo.Administrator.String() {
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

	if errSet := headerChainInfo.Administrator.SetBytes(address.Bytes()); errSet != nil {
		return nil, err
	}
	common.DealWithErr(headerChainInfo.Save(context.Storage()))
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

	headerChainInfo, err := definition.GetMergeMiningInfoVariableVariable(context.Storage())
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

	if sendBlock.Address.String() != headerChainInfo.Administrator.String() {
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

	headerChainInfo.CompressedTssECDSAPubKey = param.PubKey
	headerChainInfo.DecompressedTssECDSAPubKey = newDecompressedPubKey

	common.DealWithErr(headerChainInfo.Save(context.Storage()))
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
