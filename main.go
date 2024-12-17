/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"os"

	"github.com/AllFi/scoutup/deploy"
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

	// Subcommands
	// app.Commands = []*cli.Command{
	// 	{
	// 		Name:   config.ForkCommandName,
	// 		Usage:  "Locally fork a network in the superchain registry",
	// 		Flags:  append(config.ForkCLIFlags(envVarPrefix), baseFlags...),
	// 		Action: cliapp.LifecycleCmd(SupersimMain),
	// 	},
	// 	{
	// 		Name:   config.DocsCommandName,
	// 		Usage:  "Display available docs links",
	// 		Action: cliapp.LifecycleCmd(ScoutupMain),
	// 	},
	// }

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Crit("Application Failed", "err", err)
	}
}

func ScoutupMain(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	config1 := deploy.BlockscoutConfig{
		Name:         "Potato Chain",
		FrontendPort: 3001,
		BackendPort:  4001,
		PostgresPort: 7433,
		RpcUrl:       "http://host.docker.internal:8545/",
		FirstBlock:   5,
	}
	config2 := deploy.BlockscoutConfig{
		Name:         "Carrot Chain",
		FrontendPort: 3002,
		BackendPort:  4002,
		PostgresPort: 7434,
		RpcUrl:       "http://host.docker.internal:9545/",
		FirstBlock:   1,
	}
	config3 := deploy.BlockscoutConfig{
		Name:         "Tomato Chain",
		FrontendPort: 3003,
		BackendPort:  4003,
		PostgresPort: 7435,
		RpcUrl:       "http://host.docker.internal:9546/",
		FirstBlock:   1,
	}
	configs := []deploy.BlockscoutConfig{config1, config2, config3}
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.DefaultCLIConfig())
	deployer := deploy.New(log, closeApp, configs)
	return deployer, nil
}
