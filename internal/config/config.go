package config

type Config struct {
	RpcUrl string
}

func Default() *Config {
	return &Config{
		RpcUrl: "http://localhost:26657",
	}
}
