package config

type BlockscoutConfig struct {
	Name       string
	RpcUrl     string
	FirstBlock uint64

	FrontendPort uint64
	BackendPort  uint64
	PostgresPort uint64
}
