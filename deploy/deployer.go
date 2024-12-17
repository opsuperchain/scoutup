package deploy

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
)

type BlockscoutDeployer struct {
	instances []*Blockscout

	log      log.Logger
	closeApp context.CancelCauseFunc
}

func New(log log.Logger, closeApp context.CancelCauseFunc, configs []BlockscoutConfig) *BlockscoutDeployer {
	instances := []*Blockscout{}
	for _, config := range configs {
		instance := NewBlockscout(log, closeApp, config)
		instances = append(instances, instance)
	}
	return &BlockscoutDeployer{instances: instances, log: log, closeApp: closeApp}
}

func (b *BlockscoutDeployer) Start(ctx context.Context) error {
	for _, instance := range b.instances {
		if err := instance.Start(ctx); err != nil {
			b.log.Error("Failed to start Blockscout instance", "err", err)
			return err
		}
	}
	return nil
}

func (b *BlockscoutDeployer) Stop(ctx context.Context) error {
	for _, instance := range b.instances {
		instance.Stop(ctx)
	}
	b.log.Info("Stopped ALL Blockscout instances")
	return nil
}

// no-op dead code in the cliapp lifecycle
func (b *BlockscoutDeployer) Stopped() bool {
	return false
}
