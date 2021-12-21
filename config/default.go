package config

var (
	DefaultLogConfig = Log{
		Stdout: stdout{
			Enable: true,
			Level:  4,
		},
	}
	DefaultServerConf = ServerConf{
		Port:  8080,
		Users: map[string]string{},
	}
)
