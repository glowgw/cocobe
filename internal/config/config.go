package config

type Config struct {
	RpcUrl   string
	GrpcUrl  string
	Insecure bool
}

func DefaultConfig() Config {
	return Config{
		RpcUrl:   "http://localhost:26657",
		GrpcUrl:  "localhost:9090",
		Insecure: true,
	}
}
