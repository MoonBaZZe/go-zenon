package embedded

import (
	"github.com/inconshreveable/log15"
	"github.com/zenon-network/go-zenon/chain"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
	"github.com/zenon-network/go-zenon/zenon"
)

type MergeMiningApi struct {
	chain chain.Chain
	log   log15.Logger
}

func NewMergeMiningApi(z zenon.Zenon) *MergeMiningApi {
	return &MergeMiningApi{
		chain: z.Chain(),
		log:   common.RPCLogger.New("module", "embedded_merge_mining_api"),
	}
}

func (a *MergeMiningApi) GetMergeMiningInfo() (*definition.MergeMiningInfoVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetMergeMiningInfoVariableVariable(context.Storage())
}

func (a *MergeMiningApi) GetHeaderChainInfo() (*definition.HeaderChainInfoVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetHeaderChainInfoVariableVariable(context.Storage())
}

func (a *MergeMiningApi) GetShareChainInfo(id uint32) (*definition.ShareChainInfoVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetShareChainInfoVariableVariable(context.Storage(), id)
}

func (a *MergeMiningApi) GetBlockHeader(hash types.Hash) (*definition.BlockHeaderVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetBlockHeaderVariable(context.Storage(), hash)
}

func (a *MergeMiningApi) GetSecurityInfo() (*definition.SecurityInfoVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	security, err := definition.GetSecurityInfoVariable(context.Storage())
	if err != nil {
		return nil, err
	}

	return security, nil
}

func (a *MergeMiningApi) GetTimeChallengesInfoMergeMining() (*TimeChallengesList, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	ans := make([]*definition.TimeChallengeInfo, 0)
	// todo add the method that adds a new share chain
	methods := []string{"NominateGuardians", "ChangeTssECDSAPubKey", "ChangeAdministrator", "SetShareChain"}

	for _, m := range methods {
		timeC, err := definition.GetTimeChallengeInfoVariable(context.Storage(), m)
		if err != nil {
			return nil, err
		}
		if timeC != nil {
			ans = append(ans, timeC)
		}
	}

	return &TimeChallengesList{
		Count: len(ans),
		List:  ans,
	}, nil
}
