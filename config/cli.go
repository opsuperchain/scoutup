package config

import "github.com/urfave/cli/v2"

const (
	Anvil            = "anvil"
	Supersim         = "supersim"
	SupersimAdminRpc = "supersim.admin.rpc"
)

func BaseCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  Anvil,
			Value: false,
			Usage: "Runs single blockscout instance for default anvil config",
		},
		&cli.BoolFlag{
			Name:  Supersim,
			Value: false,
			Usage: "Fetches supersim config and runs blockscout instances for each chain",
		},
		&cli.StringFlag{
			Name:  SupersimAdminRpc,
			Value: "http://localhost:8420",
			Usage: "Admin RPC URL for supersim",
		},
	}
}
