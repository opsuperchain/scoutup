package blockscout

import (
	"context"

	"github.com/AllFi/scoutup/config"
	"github.com/ethereum/go-ethereum/log"
)

type Orchestrator struct {
	instances []*Instance

	log      log.Logger
	closeApp context.CancelCauseFunc
}

func NewOrchestrator(log log.Logger, closeApp context.CancelCauseFunc, configs []*config.BlockscoutConfig) (*Orchestrator, error) {
	globalWorkspace, err := createGlobalWorkspace()
	if err != nil {
		return nil, err
	}

	instances := []*Instance{}
	for _, config := range configs {
		instance := NewInstance(log, closeApp, config, globalWorkspace)
		instances = append(instances, instance)
	}
	return &Orchestrator{instances: instances, log: log, closeApp: closeApp}, nil
}

func (b *Orchestrator) Start(ctx context.Context) error {
	for _, instance := range b.instances {
		if err := instance.Start(ctx); err != nil {
			b.log.Error("Failed to start Blockscout instance", "err", err)
			return err
		}
	}
	return nil
}

func (b *Orchestrator) Stop(ctx context.Context) error {
	for _, instance := range b.instances {
		instance.Stop(ctx)
	}
	b.log.Info("Stopped ALL Blockscout instances")
	return nil
}

// no-op dead code in the cliapp lifecycle
func (b *Orchestrator) Stopped() bool {
	return false
}
