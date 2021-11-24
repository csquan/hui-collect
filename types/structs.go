package types

import "time"

type Base struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseTask struct {
	State int
	Message string
}


type PartReBalanceTask struct {
	*Base
	*BaseTask
	params string // used for create sub tasks
}
