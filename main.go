/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"os"

	"github.com/AllFi/scoutup/blockscout"
	"github.com/AllFi/scoutup/supersim"
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
	// network := config.NetworkConfig{
	// 	Chains: []*config.ChainConfig{
	// 		{
	// 			Name:       "L1",
	// 			RpcUrl:     "http://host.docker.internal:8545/",
	// 			FirstBlock: 0,
	// 		},
	// 		{
	// 			Name:       "OPChainA",
	// 			RpcUrl:     "http://host.docker.internal:9545/",
	// 			FirstBlock: 0,
	// 			OPConfig: &config.OPConfig{
	// 				L1RpcUrl:               "http://host.docker.internal:8545/",
	// 				L1SystemConfigContract: "0xFD19a33F8D757b8EA93BB2b40B1cDe946C1e1F4D",
	// 			},
	// 		},
	// 		{
	// 			Name:       "OPChainB",
	// 			RpcUrl:     "http://host.docker.internal:9546/",
	// 			FirstBlock: 0,
	// 			OPConfig: &config.OPConfig{
	// 				L1RpcUrl:               "http://host.docker.internal:8545/",
	// 				L1SystemConfigContract: "0xFb295Aa436F23BE2Bd17678Adf1232bdec02FED1",
	// 			},
	// 		},
	// 	},
	// }

	config, err := supersim.GenerateConfigForSupersim("http://127.0.0.1:8420")
	if err != nil {
		return nil, err
	}
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.NewOrchestrator(log, closeApp, config.GetBlockscoutConfigs())
}

func ScoutupClean(ctx *cli.Context) error {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	return blockscout.CleanupGlobalWorkspace(log)
}
