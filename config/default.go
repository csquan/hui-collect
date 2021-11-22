package config

var (
	DefaultDataBaseConfig = DataBaseConf{
		SqlBatch:         500,
		RetryIntervalInt: 500,
		RetryTimes:       5,
	}

	DefaultLogConfig = Log{
		Stdout: stdout{
			Enable: true,
			Level:  4,
		},
	}
)
