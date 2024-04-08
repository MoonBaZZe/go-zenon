package vm_context

import (
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

func (ctx *accountVmContext) IsMergeMiningEnforced() bool {
	active, err := ctx.momentumStore.IsSporkActive(types.MergeMiningSpork)
	common.DealWithErr(err)
	return active
}

func (ctx *accountVmContext) IsAcceleratorSporkEnforced() bool {
	active, err := ctx.momentumStore.IsSporkActive(types.AcceleratorSpork)
	common.DealWithErr(err)
	return active
}

func (ctx *accountVmContext) IsHtlcSporkEnforced() bool {
	active, err := ctx.momentumStore.IsSporkActive(types.HtlcSpork)
	common.DealWithErr(err)
	return active
}

func (ctx *accountVmContext) IsBridgeAndLiquiditySporkEnforced() bool {
	active, err := ctx.momentumStore.IsSporkActive(types.BridgeAndLiquiditySpork)
	common.DealWithErr(err)
	return active
}
