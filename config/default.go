package config

var (

	DefaultLogConfig = Log{
		Stdout: stdout{
			Enable: true,
			Level:  4,
		},
	}
)
