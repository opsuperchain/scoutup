package supersim

import (
	"fmt"

	"github.com/blockscout/scoutup/config"
	ssconfig "github.com/ethereum-optimism/supersim/config"
	"github.com/ethereum/go-ethereum/rpc"
)

func GenerateConfigForSupersim(admingRPCUrl string) (*config.NetworkConfig, error) {
	sc, err := fetchSupersimConfig(admingRPCUrl)
	if err != nil {
		return nil, err
	}
	println(sc)

	networkConfig := &config.NetworkConfig{
		Chains: []*config.ChainConfig{
			{
				Name:       sc.L1Config.Name,
				RpcUrl:     fmt.Sprintf("http://host.docker.internal:%d", sc.L1Config.Port),
				FirstBlock: 0, // TODO: fix me
			},
		},
	}

	for i, chain := range sc.L2Configs {
		chainConfig := &config.ChainConfig{
			Name:       chain.Name,
			RpcUrl:     fmt.Sprintf("http://host.docker.internal:%d", sc.L2StartingPort+uint64(i)), // TODO: fix me
			FirstBlock: 0,
			OPConfig: &config.OPConfig{
				L1RpcUrl:               fmt.Sprintf("http://host.docker.internal:%d", sc.L1Config.Port),
				L1SystemConfigContract: chain.L2Config.L1Addresses.SystemConfigProxy.String(),
			},
		}
		networkConfig.Chains = append(networkConfig.Chains, chainConfig)
	}

	return networkConfig, nil
}

func fetchSupersimConfig(admingRPCUrl string) (*ssconfig.NetworkConfig, error) {
	client, err := rpc.DialHTTP(admingRPCUrl)
	if err != nil {
		return nil, err
	}

	var config *ssconfig.NetworkConfig
	err = client.Call(&config, "admin_getConfig")
	if err != nil {
		return nil, err
	}

	return config, nil
}
