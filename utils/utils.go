package utils

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type message struct {
	TaskID   uint64
	TaskType string
	Content  string
}

func AppendMessage(taskMsg string, TaskID uint64, TaskType, Content string) (string, error) {
	if taskMsg == "" {
		taskMsg = "[]"
	}
	messages:= make([]*message,0)
	err := json.Unmarshal([]byte(taskMsg), &messages)
	if err != nil {
		logrus.Warnf("AppendMessage unmarshal err:%v, data:%s", err, taskMsg)
	}
	messages = append(messages, &message{
		TaskID:   TaskID,
		TaskType: TaskType,
		Content:  Content,
	})
	data, err := json.Marshal(messages)
	if err != nil {
		logrus.Warnf("AppendMessage marshal err:%v, data:%+v", err, messages)
	}
	return string(data), err
}
