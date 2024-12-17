package deploy

type BlockscoutDeployer struct {
	configs []BlockscoutConfig

	tempDir string
}

func New(configs []BlockscoutConfig, tempDir string) *BlockscoutDeployer {
	return &BlockscoutDeployer{configs: configs}
}

func Deploy() error {
	return nil
}
