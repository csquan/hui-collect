package part_rebalance

import (
	"bytes"
	"encoding/json"
	"html/template"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/alert"
	"github.com/starslabhq/hermes-rebalance/types"
	"github.com/starslabhq/hermes-rebalance/utils"
)

var lpAlertTemp = `
# stage: {{ .Stage }} - {{if .Suc }}suc{{else}}fail{{end}}
# re_id: {{ .RebalanceId }} 
# time: {{ .Time }}  
{{if .TxTasks}}
# txs:
	{{ with .TxTasks}}
	{{range .}}
	{{.ChainName}}:{{.Hash}}
	{{end}}
	{{end}}
{{end}}
-------------
{{ with .Msgs }}
{{ range . }}
## chain: {{ .Chain }}  
## vault: {{ .BaseTokenSymbol }}  
{{if .TxHash }}
## txhash: {{.TxHash}}
{{end}}
	{{ with .Lps }}
	{{ range . }}
	platform: {{ .Platform }}  
	{{if .Amount}}
	lpAmount: {{ .Amount }}  
	{{end}}
	base: {{ .BaseTokenSymbol }} {{ .BaseTokenAmount }}  
		{{if .QuoteTokenSymbol}}
			quote: {{ .QuoteTokenSymbol }} {{ .QuoteTokenAmount }}  
		{{end}}
	-------------
	{{- end}}
	{{end}}
{{- end}}
{{end}}
`

type lpAlert struct {
	Stage       string
	RebalanceId uint64
	Suc         bool
	Time        string
	TxTasks     []*types.TransactionTask
	Msgs        []*types.Pool
}

func formatLpAlertmsg(stage string, rid uint64, suc bool,
	msgs []*types.Pool, txTasks []*types.TransactionTask) (string, error) {
	lpAlert := &lpAlert{
		Stage:       stage,
		RebalanceId: rid,
		Suc:         suc,
		Time:        getTimeNow(),
		Msgs:        msgs,
		TxTasks:     txTasks,
	}
	temp := template.New("lps")
	temp = template.Must(temp.Parse(lpAlertTemp))
	var buf = &bytes.Buffer{}
	err := temp.Execute(buf, lpAlert)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func SendLpInfo(lpApi string, taskId uint64, stage string, suc bool, txTasks []*types.TransactionTask) {
	utils.HandleErrorWithRetryMaxTime(func() error {
		data, err := utils.GetLpData(lpApi)
		if err != nil {
			return err
		}
		return SendLpInfoWithData(data, taskId, stage, suc, txTasks)
	}, 3, time.Second)
}

func getTimeNow() string {
	l, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return time.Now().In(l).Format("2006-01-02 15:04:05")
	} else {
		logrus.Errorf("timenow err:%v", err)
		return time.Now().Format("")
	}
}

func SendLpInfoWithData(data *types.Data, taskId uint64, stage string, suc bool, txTasks []*types.TransactionTask) error {
	msgs, err := data.GetLpMsgs()
	if err != nil {
		b, _ := json.Marshal(data)
		logrus.Errorf("getLpMsgs err:%v,tid:%d,data:%s", err, taskId, b)
		return nil
	}

	c, err := formatLpAlertmsg(stage, taskId, suc, msgs, txTasks)
	if err != nil {
		logrus.Errorf("lp template err:%v,tid:%d", err, taskId)
		return nil
	}
	err = alert.Dingding.SendMessage(stage, c)
	if err != nil {
		return err
	}
	return nil
}
