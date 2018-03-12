package protoManager

import (
	"sync"
	"fmt"
	"github.com/zipper-project/zipper/common/mpool"
	"github.com/zipper-project/zipper/peer"
	"github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/types"
	"github.com/zipper-project/zipper/params"
	"github.com/zipper-project/zipper/blockchain/blocksync"
)

type ProtoManager struct {
	sync.Mutex
	wm map[uint32]*mpool.VirtualMachine
}

func NewProtoManager() *ProtoManager {
	return &ProtoManager{
		wm: make(map[uint32]*mpool.VirtualMachine),
	}
}

func (pm *ProtoManager) RegisterWorker(protocalID uint32, workers []mpool.VmWorker) error {
	pm.Lock()
	defer pm.Unlock()

	if _, ok := pm.wm[protocalID]; ok {
		return fmt.Errorf("wm: %s have beed registered", protocalID)
	}

	pm.wm[protocalID] = mpool.CreateCustomVM(workers)
	return nil
}

func (pm *ProtoManager) Handle(sendPeer *peer.Peer, msg *proto.Message) error {
	pm.Lock()
	defer pm.Unlock()

	err := pm.wm[msg.Header.ProtoID].SendWorkCleanAsync(types.NewWorkerData(sendPeer, msg))
	return err
}

func (pm *ProtoManager) InitAndRegisterWorker(ledger ) {
	pm.RegisterWorker(params.BlockSyncIdx, blocksync.NewSyncWorker())

}