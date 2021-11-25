package services

type crossSubState int

const (
	toCreateSubSub crossSubState = iota
	SubSubCreated
	fail
	suc
)

type CrossSubTaskService struct {
}

func (csub *CrossSubTaskService) addCross() (uint64, error) {
	return 0, nil
}
