package config

type Config struct {
	RpcUrl      string
	Connections int
	Clients     int
}

func Default() *Config {
	return &Config{
		RpcUrl:      "http://localhost:26657",
		Connections: 10,
		Clients:     2,
	}
}
