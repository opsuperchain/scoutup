package blockscout

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync/atomic"
	"syscall"

	"github.com/blockscout/scoutup/config"
	"github.com/blockscout/scoutup/utils"
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	config    *config.BlockscoutConfig
	log       log.Logger
	workspace string

	cmd *exec.Cmd

	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	closeApp       context.CancelCauseFunc

	stopped   atomic.Bool
	stoppedCh chan struct{}
}

func NewInstance(log log.Logger, closeApp context.CancelCauseFunc, config *config.BlockscoutConfig, globalWorkspace string) (*Instance, error) {
	workspace, err := createInstanceWorkspace(globalWorkspace)
	if err != nil {
		return nil, err
	}
	resCtx, resCancel := context.WithCancel(context.Background())
	return &Instance{
		config:         config,
		log:            log,
		workspace:      workspace,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		closeApp:       closeApp,
		stoppedCh:      make(chan struct{}, 1),
	}, nil
}

func (i *Instance) Start(ctx context.Context) error {
	i.log.Info("Starting Blockscout", "chain", i.config.Name)

	err := i.configureBlockscout()
	if err != nil {
		return err
	}

	err = i.runDockerCompose(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (i *Instance) Stop(_ context.Context) error {
	i.log.Info("Stopping Blockscout", "chain", i.config.Name)
	if i.stopped.Load() {
		return errors.New("already stopped")
	}
	if !i.stopped.CompareAndSwap(false, true) {
		return nil // someone else stopped
	}

	i.resourceCancel()
	<-i.stoppedCh
	return nil
}

// no-op dead code in the cliapp lifecycle
func (i *Instance) Stopped() bool {
	return false
}

func (i *Instance) ConfigAsString() string {
	var b strings.Builder
	fmt.Fprintf(&b, "* Chain: %v\n", i.config.Name)
	fmt.Fprintf(&b, "         Frontend:  http://127.0.0.1:%v\n", i.config.FrontendPort)
	fmt.Fprintf(&b, "         Backend:   http://127.0.0.1:%v\n", i.config.BackendPort)
	fmt.Fprintf(&b, "         DB:        http://127.0.0.1:%v\n", i.config.PostgresPort)
	fmt.Fprintf(&b, "         Workspace: %v\n", i.workspace)
	fmt.Fprintf(&b, "         Logs:	    %v\n", path.Join(i.workspace, "logs"))

	if i.config.OPConfig != nil {
		fmt.Fprintf(&b, "         Optimism L1 RPC: %v\n", i.config.OPConfig.L1RPCUrl)
		fmt.Fprintf(&b, "         Optimism L1 System Config Contract: %v\n", i.config.OPConfig.L1SystemConfigContract)
	}
	return b.String()
}

func (i *Instance) configureBlockscout() error {
	utils.PatchDotEnv(path.Join(i.workspace, "common-blockscout.env"), i.backendEnvs())
	utils.PatchDotEnv(path.Join(i.workspace, "common-frontend.env"), i.frontendEnvs())
	return nil
}

func (i *Instance) runDockerCompose(ctx context.Context) error {
	i.cmd = exec.CommandContext(i.resourceCtx, "docker", "compose", "up")
	i.cmd.Env = append(os.Environ(), i.dockerComposeEnvs()...)
	i.cmd.Cancel = func() error {
		return i.cmd.Process.Signal(syscall.SIGTERM)
	}
	i.cmd.Dir = i.workspace
	go func() {
		<-ctx.Done()
		i.resourceCancel()
	}()

	stdout, err := i.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	logFile, err := os.Create(path.Join(i.workspace, "logs"))
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			txt := scanner.Text()
			if _, err := fmt.Fprintln(logFile, txt); err != nil {
				i.log.Warn("err piping stdout to log file", "err", err)
			}
		}
	}()

	stderr, err := i.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			txt := scanner.Text()
			if _, err := fmt.Fprintln(logFile, txt); err != nil {
				i.log.Warn("err piping stdout to log file", "err", err)
			}
		}
	}()

	if err := i.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := i.cmd.Wait(); err != nil {
			if err.Error() != "exit status 130" {
				i.log.Error("Blockscout terminated with an error", "error", err)
			}
		} else {
			i.log.Info("Blockscout terminated")
		}

		err := cleanupInstanceWorkspace(i.workspace)
		if err != nil {
			i.log.Error("Failed to cleanup workspace", "error", err)
		}

		// If it stops, signal that the entire app should be closed
		i.closeApp(nil)
		i.stoppedCh <- struct{}{}
	}()

	return nil
}

func (i *Instance) dockerComposeEnvs() []string {
	return []string{
		fmt.Sprintf("DOCKER_REPO=%s", i.config.DockerRepo),
		fmt.Sprintf("FRONTEND_PORT=%d", i.config.FrontendPort),
		fmt.Sprintf("BACKEND_PORT=%d", i.config.BackendPort),
		fmt.Sprintf("POSTGRES_PORT=%d", i.config.PostgresPort),
		fmt.Sprintf("DB_CONTAINER_NAME=%s", utils.NameToContainerName("db", i.config.Name)),
		fmt.Sprintf("BACKEND_CONTAINER_NAME=%s", utils.NameToContainerName("backend", i.config.Name)),
		fmt.Sprintf("FRONTEND_CONTAINER_NAME=%s", utils.NameToContainerName("frontend", i.config.Name)),
	}
}

func (i *Instance) backendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["ETHEREUM_JSONRPC_HTTP_URL"] = i.config.RPCUrl
	envs["ETHEREUM_JSONRPC_TRACE_URL"] = i.config.RPCUrl
	envs["SUBNETWORK"] = i.config.Name
	envs["FIRST_BLOCK"] = fmt.Sprintf("%d", i.config.FirstBlock)
	envs["DATABASE_URL"] = fmt.Sprintf(
		"postgresql://blockscout:ceWb1MeLBEeOIfk65gU8EjF8@host.docker.internal:%v/blockscout", i.config.PostgresPort)
	if i.config.OPConfig != nil {
		envs["INDEXER_OPTIMISM_L1_RPC"] = i.config.OPConfig.L1RPCUrl
		envs["INDEXER_OPTIMISM_L1_SYSTEM_CONFIG_CONTRACT"] = i.config.OPConfig.L1SystemConfigContract
		envs["INDEXER_OPTIMISM_L2_BATCH_GENESIS_BLOCK_NUMBER"] = "0"
		envs["INDEXER_OPTIMISM_L2_HOLOCENE_TIMESTAMP"] = "0"
	}
	return envs
}

func (i *Instance) frontendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["NEXT_PUBLIC_API_PORT"] = fmt.Sprintf("%d", i.config.BackendPort)
	envs["NEXT_PUBLIC_NETWORK_NAME"] = i.config.Name
	envs["NEXT_PUBLIC_NETWORK_SHORT_NAME"] = i.config.Name

	if i.config.OPConfig != nil {
		envs["NEXT_PUBLIC_ROLLUP_TYPE"] = "optimistic"
		// TODO: Fix me
		envs["NEXT_PUBLIC_ROLLUP_L1_BASE_URL"] = "http://host.docker.internal:8545"
		envs["NEXT_PUBLIC_ROLLUP_L2_WITHDRAWAL_URL"] = "https://app.optimism.io/bridge/withdraw"
	}

	return envs
}
