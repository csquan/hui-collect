package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
)

type CheckReceiptService struct {
	db     types.IDB
	config *config.Config
}

func NewCheckReceiptService(db types.IDB, c *config.Config) *CheckReceiptService {
	return &CheckReceiptService{
		db:     db,
		config: c,
	}
}

func (c *CheckReceiptService) CheckReceipt(task *types.TransactionTask) (finished bool, err error) {
	receipt, err := c.handleCheckReceipt(task)
	if err != nil {
		return false, err
	}
	b, err := json.Marshal(receipt)
	if err != nil {
		return false, err
	}
	task.Receipt = string(b)
	task.State = int(types.TxSuccessState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in CheckReceipt err:%v", err)
	}
	return true, nil
}

func (c *CheckReceiptService) handleCheckReceipt(task *types.TransactionTask) (*ethtypes.Receipt, error) {
	rawTxBytes, err := hex.DecodeString(task.SignData)
	if err != nil {
		return nil, err
	}
	tx := new(ethtypes.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx)

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (c *CheckReceiptService) tgalert(task *types.TransactionTask) {

}

func (c *CheckReceiptService) Run() error {
	tasks, err := c.db.GetOpenedCheckReceiptTasks()
	if err != nil {
		return fmt.Errorf("get tasks for check receipt err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.CheckReceipt(task)
		if err == nil {
			c.tgalert(task)
		}
	}
	return nil
}

func (c CheckReceiptService) Name() string {
	return "CheckReceipt"
}
