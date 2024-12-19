package config

import "fmt"

type OPConfig struct {
	L1RPCUrl               string
	L1SystemConfigContract string
	L1BlockscoutURL        string
}

type ChainConfig struct {
	Name       string
	RPCUrl     string
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
	frontendPort := n.StartingFrontendPort
	backendPort := n.StartingBackendPort
	postgresPort := n.StartingPostgresPort

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

		if config.OPConfig != nil {
			// TODO: refactor this later
			for _, bs := range configs {
				if bs.RPCUrl == config.OPConfig.L1RPCUrl {
					config.OPConfig.L1BlockscoutURL = fmt.Sprintf("http://host.docker.internal:%v", bs.FrontendPort)
					break
				}
			}
		}

		configs = append(configs, config)
		frontendPort++
		backendPort++
		postgresPort++
	}
	return configs
}

func (n *ChainConfig) dockerRepo() string {
	if n.OPConfig != nil {
		return "blockscout-optimism"
	}
	return "blockscout"
}
