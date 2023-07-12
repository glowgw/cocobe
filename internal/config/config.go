package config

type Config struct {
	RpcUrl   string
	GrpcUrl  string
	Insecure bool
}

func Default() *Config {
	return &Config{
		RpcUrl: "http://10.15.7.100:26657",
		// RpcUrl:   "http://localhost:26657",
		GrpcUrl:  "localhost:9090",
		Insecure: true,
	}
}
