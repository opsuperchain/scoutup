package blockscout

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync/atomic"
	"syscall"

	"github.com/AllFi/scoutup/config"
	"github.com/AllFi/scoutup/utils"
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	config          *config.BlockscoutConfig
	log             log.Logger
	globalWorkspace string

	cmd *exec.Cmd

	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	closeApp       context.CancelCauseFunc

	stopped   atomic.Bool
	stoppedCh chan struct{}
}

func NewInstance(log log.Logger, closeApp context.CancelCauseFunc, config *config.BlockscoutConfig, globalWorkspace string) *Instance {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &Instance{
		config:          config,
		log:             log,
		globalWorkspace: globalWorkspace,
		resourceCtx:     resCtx,
		resourceCancel:  resCancel,
		closeApp:        closeApp,
		stoppedCh:       make(chan struct{}, 1),
	}
}

func (b *Instance) Start(ctx context.Context) error {
	b.log.Info("Starting Blockscout instance")

	tempDir, err := createInstanceWorkspace(b.globalWorkspace)
	if err != nil {
		return err
	}

	err = b.configureBlockscout(tempDir)
	if err != nil {
		return err
	}

	err = b.runDockerCompose(ctx, tempDir)
	if err != nil {
		return err
	}

	return nil
}

func (b *Instance) Stop(_ context.Context) error {
	b.log.Info("Stopping Blockscout instance")
	if b.stopped.Load() {
		return errors.New("already stopped")
	}
	if !b.stopped.CompareAndSwap(false, true) {
		return nil // someone else stopped
	}

	b.resourceCancel()
	<-b.stoppedCh
	return nil
}

// no-op dead code in the cliapp lifecycle
func (b *Instance) Stopped() bool {
	return false
}

func (b *Instance) configureBlockscout(tempDir string) error {
	b.log.Info("Configuring Blockscout")
	utils.PatchDotEnv(path.Join(tempDir, "common-blockscout.env"), b.backendEnvs())
	utils.PatchDotEnv(path.Join(tempDir, "common-frontend.env"), b.frontendEnvs())
	return nil
}

func (b *Instance) runDockerCompose(ctx context.Context, tempDir string) error {
	b.log.Info("Starting Blockscout with docker-compose")
	b.cmd = exec.CommandContext(b.resourceCtx, "docker", "compose", "up")
	b.cmd.Env = append(os.Environ(), b.dockerComposeEnvs()...)
	b.cmd.Cancel = func() error {
		return b.cmd.Process.Signal(syscall.SIGTERM)
	}
	b.cmd.Dir = tempDir
	go func() {
		<-ctx.Done()
		b.resourceCancel()
	}()

	if err := b.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := b.cmd.Wait(); err != nil {
			if err.Error() != "exit status 130" {
				b.log.Error("blockscout terminated with an error", "error", err)
			}
		} else {
			b.log.Info("blockscout terminated")
		}

		err := cleanupInstanceWorkspace(tempDir)
		if err != nil {
			b.log.Error("Failed to cleanup workspace", "error", err)
		}

		// If it stops, signal that the entire app should be closed
		b.closeApp(nil)
		b.stoppedCh <- struct{}{}
	}()

	return nil
}

func (b *Instance) dockerComposeEnvs() []string {
	return []string{
		fmt.Sprintf("FRONTEND_PORT=%d", b.config.FrontendPort),
		fmt.Sprintf("BACKEND_PORT=%d", b.config.BackendPort),
		fmt.Sprintf("POSTGRES_PORT=%d", b.config.PostgresPort),
		fmt.Sprintf("DB_CONTAINER_NAME=%s", utils.NameToContainerName("db", b.config.Name)),
		fmt.Sprintf("BACKEND_CONTAINER_NAME=%s", utils.NameToContainerName("backend", b.config.Name)),
		fmt.Sprintf("FRONTEND_CONTAINER_NAME=%s", utils.NameToContainerName("frontend", b.config.Name)),
	}
}

func (b *Instance) backendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["ETHEREUM_JSONRPC_HTTP_URL"] = b.config.RpcUrl
	envs["ETHEREUM_JSONRPC_TRACE_URL"] = b.config.RpcUrl
	envs["SUBNETWORK"] = b.config.Name
	envs["FIRST_BLOCK"] = fmt.Sprintf("%d", b.config.FirstBlock)
	envs["DATABASE_URL"] = fmt.Sprintf(
		"postgresql://blockscout:ceWb1MeLBEeOIfk65gU8EjF8@host.docker.internal:%v/blockscout", b.config.PostgresPort)
	return envs
}

func (b *Instance) frontendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["NEXT_PUBLIC_API_PORT"] = fmt.Sprintf("%d", b.config.BackendPort)
	envs["NEXT_PUBLIC_NETWORK_NAME"] = b.config.Name
	envs["NEXT_PUBLIC_NETWORK_SHORT_NAME"] = b.config.Name
	return envs
}
