package config

type OPConfig struct {
	L1RpcUrl               string
	L1SystemConfigContract string
}

type ChainConfig struct {
	Name       string
	RpcUrl     string
	FirstBlock uint64
	OPConfig   *OPConfig
}

type InstanceConfig struct {
	DockerRepo   string
	FrontendPort uint64
	BackendPort  uint64
	PostgresPort uint64
}

type BlockscoutConfig struct {
	*ChainConfig
	*InstanceConfig
}

type NetworkConfig struct {
	Chains               []*ChainConfig
	StartingFrontendPort uint64
	StartingBackendPort  uint64
	StartingPostgresPort uint64
}

func (n *NetworkConfig) GetBlockscoutConfigs() []*BlockscoutConfig {
	frontendPort := n.startingFrontendPort()
	backendPort := n.startingBackendPort()
	postgresPort := n.startingPostgresPort()

	configs := []*BlockscoutConfig{}
	for _, chain := range n.Chains {
		config := &BlockscoutConfig{
			ChainConfig: chain,
			InstanceConfig: &InstanceConfig{
				FrontendPort: frontendPort,
				BackendPort:  backendPort,
				PostgresPort: postgresPort,
				DockerRepo:   chain.dockerRepo(),
			},
		}
		configs = append(configs, config)
		frontendPort++
		backendPort++
		postgresPort++
	}
	return configs
}

func (n *NetworkConfig) startingFrontendPort() uint64 {
	if n.StartingFrontendPort == uint64(0) {
		return 3000
	}
	return n.StartingFrontendPort
}

func (n *NetworkConfig) startingBackendPort() uint64 {
	if n.StartingBackendPort == uint64(0) {
		return 4000
	}
	return n.StartingBackendPort
}

func (n *NetworkConfig) startingPostgresPort() uint64 {
	if n.StartingPostgresPort == uint64(0) {
		return 7432
	}
	return n.StartingPostgresPort
}

func (n *ChainConfig) dockerRepo() string {
	if n.OPConfig != nil {
		return "blockscout-optimism"
	}
	return "blockscout"
}
