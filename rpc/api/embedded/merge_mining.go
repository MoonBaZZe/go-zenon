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

func (a *MergeMiningApi) GetHeaderChainInfo() (*definition.HeaderChainInfoVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetHeaderChainInfoVariableVariable(context.Storage())
}

func (a *MergeMiningApi) GetBlockHeader(hash types.Hash) (*definition.BlockHeaderVariable, error) {
	_, context, err := api.GetFrontierContext(a.chain, types.MergeMiningContract)
	if err != nil {
		return nil, err
	}

	return definition.GetBlockHeaderVariable(context.Storage(), hash)
}
