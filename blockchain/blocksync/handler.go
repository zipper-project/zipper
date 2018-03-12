package blocksync

import (
	"errors"
	"fmt"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/ledger"
	"github.com/zipper-project/zipper/peer"
	msgProto "github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/types"
)

type SyncWorker struct {
	ledger *ledger.Ledger
	bc     *blockchain.Blockchain
}

func (worker *SyncWorker) VmJob(data interface{}) (interface{}, error) {
	workerData := data.(types.WorkerData)
	msg := workerData.GetMsg()
	switch proto.MsgType(msg.Header.MsgID) {
	case proto.MsgType_BC_OnStatusMSg:
		worker.OnStatus(workerData)
	case proto.MsgType_BC_GetBlocksMsg:
		worker.OnGetBlocks(workerData)
	case proto.MsgType_BC_GetInvMsg:
		worker.OnGetInv(workerData)
	case proto.MsgType_BC_GetDataMsg:
		worker.OnGetData(workerData)
	case proto.MsgType_BC_OnBlockMsg:
		worker.OnBlock(workerData)
	case proto.MsgType_BC_OnTransactionMsg:
		worker.OnTransaction(workerData)
	default:
		return nil, errors.New("Not support this type of message")
	}

	return nil, nil
}

func (worker *SyncWorker) VmReady() bool {
	return true
}

func NewSyncWorker(ledger *ledger.Ledger, bc *blockchain.Blockchain) *SyncWorker {
	return &SyncWorker{
		ledger: ledger,
		bc:     bc,
	}
}

func SetSyncWorkers(workerNums int, ledger *ledger.Ledger, bc *blockchain.Blockchain) []*SyncWorker {
	var txWorkers []*SyncWorker
	for i := 0; i < workerNums; i++ {
		txWorkers = append(txWorkers, NewSyncWorker(ledger, bc))
	}

	return txWorkers
}

func (worker *SyncWorker) OnStatus(workerData types.WorkerData) {

}

func (worker *SyncWorker) OnGetBlocks(workerData types.WorkerData) {
	getBlocksMsg := proto.GetBlocksMsg{}
	if err := getBlocksMsg.UnmarshalMsg(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetBlocksMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	//check all hash for get, and get the starting hash
	var startHash string
	for _, h := range getBlocksMsg.LocatorHashes {
		_, err := worker.ledger.GetBlockByHash([]byte(h))
		if err != nil {
			break
		}

		startHash = h
	}

	hash := crypto.NewHash([]byte(startHash))
	hashes := []string{}
	var err error
	for {
		hash, err = worker.bc.GetNextBlockHash(hash)
		if err != nil || hash.Equal(crypto.Hash{}) {
			break
		} else {
			hashes = append(hashes, hash.String())
		}
	}

	if len(hashes) > 0 {
		getInvMsg := &proto.GetInvMsg{
			Type:  proto.InvType_block,
			Hashs: hashes,
		}

		worker.SendMsg(workerData.GetSendPeer(), 1, 1, getInvMsg)
	}
}

func (worker *SyncWorker) OnGetInv(workerData types.WorkerData) {
	if worker.bc.Synced() {
		return
	}

	getInvMsg := &proto.GetInvMsg{}
	if err := getInvMsg.UnmarshalMsg(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetInvMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	hashes := []string{}
	switch getInvMsg.Type {
	case proto.InvType_block:
		for _, h := range getInvMsg.Hashs {
			if tx, _ := worker.bc.GetTransaction(crypto.NewHash([]byte(h))); tx != nil {
				hashes = append(hashes, h)
			}
		}
	case proto.InvType_transaction:
		for _, h := range getInvMsg.Hashs {
			if block, _ := worker.ledger.GetBlockByHash([]byte(h)); block == nil {
				hashes = append(hashes, h)
			}
		}

	default:
		log.Errorf("Not support this type: %+v in GetInvMsg", getInvMsg.Type)
	}

	if len(hashes) > 0 {
		getDataMsg := &proto.GetDataMsg{}
		getDataMsg.InvList[0].Hashs = hashes

		worker.SendMsg(workerData.GetSendPeer(), 1, 1, getDataMsg)
	}
}

func (worker *SyncWorker) OnGetData(workerData types.WorkerData) {
	getDataMsg := &proto.GetDataMsg{}
	if err := getDataMsg.UnmarshalMsg(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetDataMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	for _, inv := range getDataMsg.InvList {
		switch inv.Type {
		case proto.InvType_block:
			for _, h := range inv.Hashs {
				if header, err := worker.ledger.GetBlockByHash([]byte(h)); err == nil && header != nil {
					txs, err := worker.ledger.GetTxsByBlockHash([]byte(h), 100)
					if err != nil {
						log.Errorf("Can't GetTxsByBlockHash, err: %+v", err)
						break
					}

					blockMsg := &proto.OnBlockMsg{
						Block: &proto.Block{
							Header:       header,
							Transactions: txs,
						},
					}

					worker.SendMsg(workerData.GetSendPeer(), 1, 1, blockMsg)
				}
			}
		case proto.InvType_transaction:
			for _, h := range inv.Hashs {
				if tx, _ := worker.bc.GetTransaction(crypto.NewHash([]byte(h))); tx != nil {
					txMsg := &proto.OnTransactionMsg{
						Transaction: tx,
					}

					worker.SendMsg(workerData.GetSendPeer(), 1, 1, txMsg)
				}
			}
		default:
			log.Errorf("Not support this type: %+v in OnGetDataMsg", inv.Type)
		}
	}

}

func (worker *SyncWorker) OnBlock(workerData types.WorkerData) {
	blockMsg := &proto.OnBlockMsg{}
	if err := blockMsg.UnmarshalMsg(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid OnBlockMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	if worker.bc.CurrentHeight()+1 < blockMsg.Block.Header.Height {
		getBlocksMsg := &proto.GetBlocksMsg{
			LocatorHashes: []string{worker.bc.CurrentBlockHash().String()},
			HashStop:      crypto.Hash{}.String(),
		}

		worker.SendMsg(workerData.GetSendPeer(), 1, 1, getBlocksMsg)
	} else if worker.bc.Relay(blockMsg.Block) {
		if !worker.bc.Started() && worker.bc.CurrentHeight() == worker.expectedHeight {
			worker.bc.StartService()
		}
	}
}

func (worker *SyncWorker) OnTransaction(workerData types.WorkerData) {
	txMsg := &proto.OnTransactionMsg{}
	if err := txMsg.UnmarshalMsg(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid OnTransactionMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	worker.bc.Relay(txMsg.Transaction)
}

func (worker *SyncWorker) SendMsg(peer *peer.Peer, protoID, msgID uint32, imsg proto.IMsg) error {
	msg := &msgProto.Message{}
	msg.Header.ProtoID = protoID
	msg.Header.MsgID = msgID

	imsgByte, err := imsg.MarshalMsg()
	if err != nil {
		return fmt.Errorf("to create invalid msg: %+v", imsg)
	}

	msg.Payload = imsgByte
	if data, err := msg.MarshalMsg(); err != nil {
		err = peer.SendMsg(data)
		if err != nil {
			return fmt.Errorf("can't send msg[%+v] to peer[%+v]", msg, peer.ID)
		}
	}

	return nil
}
