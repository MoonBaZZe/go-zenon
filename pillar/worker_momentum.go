package pillar

import (
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/consensus"
)

func (w *worker) generateMomentum(e consensus.ProducerEvent) (*nom.MomentumTransaction, error) {
	insert := w.chain.AcquireInsert("momentum-generator")
	defer insert.Unlock()

	store := w.chain.GetFrontierMomentumStore()

	feeSporkActive, err := store.IsSporkActive(types.FeeSpork)
	common.DealWithErr(err)
	blocks := w.chain.GetNewMomentumContent(feeSporkActive)

	previousMomentum, err := store.GetFrontierMomentum()
	if err != nil {
		return nil, err
	}

	var content nom.MomentumContent
	if feeSporkActive {
		content = nom.NewMomentumContentFeeSporkActive(blocks)
	} else {
		content = nom.NewMomentumContent(blocks)
	}

	m := &nom.Momentum{
		ChainIdentifier: w.chain.ChainIdentifier(),
		PreviousHash:    previousMomentum.Hash,
		Height:          previousMomentum.Height + 1,
		TimestampUnix:   uint64(e.StartTime.Unix()),
		Content:         content,
		Version:         uint64(1),
	}
	m.EnsureCache()
	return w.supervisor.GenerateMomentum(&nom.DetailedMomentum{
		Momentum:      m,
		AccountBlocks: blocks,
	}, w.coinbase.Signer)
}
