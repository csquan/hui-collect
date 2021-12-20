package config

var (
	DefaultLogConfig = Log{
		Stdout: stdout{
			Enable: true,
			Level:  4,
		},
	}
	DefaultAPIConf = APIConf{
		Port:  8080,
		Users: map[string]string{},
	}
)
