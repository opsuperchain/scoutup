package deploy

type BlockscoutConfig struct {
	Name         string
	FrontendPort uint64
	RpcUrl       string
	FirstBlock   uint64
}
