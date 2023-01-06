package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/HuiCollect/config"
	"github.com/ethereum/HuiCollect/types"
	"github.com/ethereum/HuiCollect/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"math/big"
	"strconv"
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
	task.TxHash = hash //广播的交易hash
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
	gasPrice, err := strconv.ParseInt(task.GasPrice, 10, 64)
	if err != nil {
		return "", err
	}

	to := common.HexToAddress(task.To)

	value, err := hexutil.DecodeBig(task.Value)
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		return "", err
	}

	b, err := hex.DecodeString(task.InputData[2:])
	if err != nil {
		return "", err
	}
	tx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    task.Nonce,
		GasPrice: big.NewInt(gasPrice),
		Gas:      8000000,
		To:       &to,
		Value:    value,
		Data:     b,
	})

	var sigData types.SigData
	err = json.Unmarshal([]byte(task.Sig), &sigData)

	remoteSig, err := hex.DecodeString(sigData.Signature)

	signer := ethTypes.NewEIP155Signer(big.NewInt(int64(task.ChainId)))
	sigedTx, err := tx.WithSignature(signer, remoteSig)

	if err != nil {
		return "", err
	}

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), sigedTx)
	if err != nil {
		fmt.Printf(" send err:" + err.Error() + "\n")
		task.Error = err.Error()
		return "", err
	}

	fmt.Printf("tx sent: %s", sigedTx.Hash())

	return sigedTx.Hash().Hex(), nil

}

func (c *BroadcastService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createBoradcastMsg(task)
	if err != nil {
		logrus.Errorf("create assembly msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
	}

	bot, err := tgbot.NewBot("5904746042:AAGjBMN_ahQ0uavSCakrEFUN7RV2Q8oDY4I")
	if err != nil {
		logrus.Fatal(err)
	}
	err = bot.SendMsg(-1001731474163, msg)
	if err != nil {
		logrus.Fatal(err)
	}
}
func createBoradcastMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易广播完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("SignHash: %v\n\n", task.SignHash))
	buffer.WriteString(fmt.Sprintf("TxHash: %v\n\n", task.TxHash))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
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
			//c.tgAlert(task)
		}
	}
	return nil
}

func (c BroadcastService) Name() string {
	return "Broadcast"
}
