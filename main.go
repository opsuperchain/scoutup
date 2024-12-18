/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"os"

	"github.com/AllFi/scoutup/blockscout"
	"github.com/AllFi/scoutup/config"
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

	//baseFlags := append(config.BaseCLIFlags(envVarPrefix), logFlags...)

	// Vanilla mode has no specific flags for now
	//app.Flags = baseFlags

	//Subcommands
	app.Commands = []*cli.Command{
		{
			Name:   "clean",
			Usage:  "Clean up all all containers and temporary files",
			Action: ScoutupClean,
		},
	}

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Crit("Application Failed", "err", err)
	}
}

func ScoutupMain(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	network := config.NetworkConfig{
		Chains: []*config.ChainConfig{
			{
				Name:       "Potato Chain",
				RpcUrl:     "http://host.docker.internal:8545/",
				FirstBlock: 5,
			},
			{
				Name:       "Carrot Chain",
				RpcUrl:     "http://host.docker.internal:9545/",
				FirstBlock: 1,
			},
			{
				Name:       "Tomato Chain",
				RpcUrl:     "http://host.docker.internal:9546/",
				FirstBlock: 1,
			},
		},
	}
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.NewOrchestrator(log, closeApp, network.GetBlockscoutConfigs())
}

func ScoutupClean(ctx *cli.Context) error {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.CleanupGlobalWorkspace(log)
}
