package upgrade

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

var (
	VersionToBeSwitched Version
)

func (keeper Keeper) GetVersionToBeSwitched() *Version {
	return &VersionToBeSwitched
}

func (keeper Keeper) RegisterVersionToBeSwitched(store sdk.KVStore, router bam.Router) {
	currentVersion := keeper.GetCurrentVersionByStore(store)

	if currentVersion == nil { // waiting to create the genesis version
		return
	}

	modulelist := NewModuleLifeTimeList()
	handlerList := router.RouteTable()

	for _, handler := range handlerList {
		hs := strings.Split(handler, "/")

		stores := strings.Split(hs[1], ":")
		modulelist = modulelist.BuildModuleLifeTime(0, hs[0], stores)
	}

	VersionToBeSwitched = NewVersion(currentVersion.Id+1, 0, 0, modulelist)
}

func (k Keeper) SetDoingSwitch(ctx sdk.Context, doing bool) {
	kvStore := ctx.KVStore(k.storeKey)

	var bytes []byte
	if doing {
		bytes = []byte{byte(1)}
	} else {
		bytes = []byte{byte(0)}
	}
	kvStore.Set(GetDoingSwitchKey(), bytes)
}

func (k Keeper) GetDoingSwitch(ctx sdk.Context) bool {
	kvStore := ctx.KVStore(k.storeKey)

	bytes := kvStore.Get(GetDoingSwitchKey())
	if len(bytes) == 1 {
		return bytes[0] == byte(1)
	}

	return false
}

func (k Keeper) DoSwitchBegin(ctx sdk.Context) {
	k.SetDoingSwitch(ctx, true)
}

func (k Keeper) DoSwitchEnd(ctx sdk.Context) {
	VersionToBeSwitched.ProposalID = k.GetCurrentProposalID(ctx)
	VersionToBeSwitched.Start = ctx.BlockHeight()

	k.AddNewVersion(ctx, VersionToBeSwitched)

	k.SetDoingSwitch(ctx, false)
	k.SetCurrentProposalID(ctx, -1)
}
