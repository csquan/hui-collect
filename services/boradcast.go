package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

type BroadcastService struct {
	db     types.IDB
	config *config.Config
}

func NewBoradcastService(db types.IDB, c *config.Config) *BroadcastService {
	return &BroadcastService{
		db:     db,
		config: c,
	}
}

func (c *BroadcastService) BroadcastTx(task *types.TransactionTask) (finished bool, err error) {
	hash, err := c.handleBroadcastTx(task)
	if err != nil {
		return false, err
	}
	task.Hash = hash
	task.State = int(types.TxBroadcastState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in BroadcastTx err:%v", err)
	}
	return true, nil
}

func (c *BroadcastService) handleBroadcastTx(task *types.TransactionTask) (string, error) {
	rawTxBytes, err := hex.DecodeString(task.SignData)
	if err != nil {
		log.Fatal(err)
	}
	tx := new(ethtypes.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx)

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", err
	}

	fmt.Printf("tx sent: %s", tx.Hash().Hex())
	return tx.Hash().Hex(), nil
}

func (c *BroadcastService) tgalert(task *types.TransactionTask) {

}

func (c *BroadcastService) Run() error {
	tasks, err := c.db.GetOpenedBroadcastTasks()
	if err != nil {
		return fmt.Errorf("get tasks for broadcast err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.BroadcastTx(task)
		if err == nil {
			c.tgalert(task)
		}
	}
	return nil
}

func (c BroadcastService) Name() string {
	return "Broadcast"
}
