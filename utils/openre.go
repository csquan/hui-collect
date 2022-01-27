package utils

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/starslabhq/hermes-rebalance/types"
)

func GetLpData(url string) (lpList *types.Data, err error) {
	data, err := DoRequest(url, "GET", nil)
	if err != nil {
		logrus.Errorf("request lp err:%v,body:%s", err, data)
		return
	}
	lpResponse := &types.LPResponse{}
	if err = json.Unmarshal(data, lpResponse); err != nil {
		logrus.Errorf("unmarshar lpResponse err:%v,body:%s", err, data)
		return
	}
	if lpResponse.Code != 200 {
		err = fmt.Errorf("lpResponse code not 200, msg:%s", lpResponse.Msg)
		logrus.Error(err)
		return
	}
	lpList = lpResponse.Data
	return
}
