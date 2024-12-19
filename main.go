/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"os"

	"github.com/blockscout/scoutup/blockscout"
	"github.com/blockscout/scoutup/config"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	"github.com/ethereum-optimism/optimism/op-service/ctxinterrupt"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "scoutup"
	app.Usage = "Dev tool for running local blockscout instances"
	app.Description = "Dev tool for running local blockscout instances"
	app.Action = cliapp.LifecycleCmd(ScoutupMain)
	app.Flags = config.BaseCLIFlags()

	//Subcommands
	app.Commands = []*cli.Command{
		{
			Name:   "clean",
			Usage:  "Cleans up all containers and temporary files",
			Action: ScoutupClean,
		},
	}

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Crit("Application Failed", "err", err)
	}
}

func ScoutupMain(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	var networkConfig *config.NetworkConfig
	if ctx.Bool(config.Supersim) {
		var err error
		networkConfig, err = config.PrepareSupersimConfig(ctx.String(config.SupersimAdminRpc))
		if err != nil {
			return nil, err
		}
	} else {
		networkConfig = config.PrepareDefaultAnvilConfig()
	}
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.NewOrchestrator(log, closeApp, networkConfig.GetBlockscoutConfigs())
}

func ScoutupClean(ctx *cli.Context) error {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.CleanupGlobalWorkspace(log)
}
