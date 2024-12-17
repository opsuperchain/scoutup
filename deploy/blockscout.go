package deploy

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync/atomic"
	"syscall"

	"github.com/ethereum/go-ethereum/log"
)

const (
	DockerComposeCommand = "docker compose"
)

var (
	//go:embed docker/docker-compose.yml
	dockerComposeYml []byte

	//go:embed docker/common-blockscout.env
	commonBlockscoutEnv []byte

	//go:embed docker/common-frontend.env
	commonFrontendEnv []byte
)

type Blockscout struct {
	config BlockscoutConfig
	log    log.Logger

	cmd *exec.Cmd

	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	closeApp       context.CancelCauseFunc

	stopped   atomic.Bool
	stoppedCh chan struct{}

	cleanupTasks []func()
}

func NewBlockscout(log log.Logger, closeApp context.CancelCauseFunc, config BlockscoutConfig) *Blockscout {
	resCtx, resCancel := context.WithCancel(context.Background())
	return &Blockscout{
		config:         config,
		log:            log,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		closeApp:       closeApp,
		stoppedCh:      make(chan struct{}, 1),
	}
}

func (b *Blockscout) Start(ctx context.Context) error {
	b.log.Info("Starting Blockscout instance")

	tempDir, err := b.setupTempDir()
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

func (b *Blockscout) Stop(_ context.Context) error {
	b.log.Info("Stopping Blockscout instance")
	if b.stopped.Load() {
		return errors.New("already stopped")
	}
	if !b.stopped.CompareAndSwap(false, true) {
		return nil // someone else stopped
	}

	b.resourceCancel()
	<-b.stoppedCh
	b.executeCleanup()
	return nil
}

// no-op dead code in the cliapp lifecycle
func (b *Blockscout) Stopped() bool {
	return false
}

func (b *Blockscout) setupTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "blockscout")
	if err != nil {
		return "", err
	}
	b.log.Info("Creating temporary directory", "path", tempDir)

	b.registerCleanupTask(func() {
		os.RemoveAll(tempDir)
	})

	files := map[string][]byte{
		"docker-compose.yml":    dockerComposeYml,
		"common-blockscout.env": commonBlockscoutEnv,
		"common-frontend.env":   commonFrontendEnv,
	}

	for name, content := range files {
		err := os.WriteFile(path.Join(tempDir, name), content, 0644)
		if err != nil {
			return "", err
		}
	}
	return tempDir, nil
}

func (b *Blockscout) configureBlockscout(tempDir string) error {
	b.log.Info("Configuring Blockscout")
	patchDotEnv(path.Join(tempDir, "common-blockscout.env"), b.backendEnvs())
	patchDotEnv(path.Join(tempDir, "common-frontend.env"), b.frontendEnvs())
	return nil
}

func (b *Blockscout) runDockerCompose(ctx context.Context, tempDir string) error {
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
			b.log.Error("blockscout terminated with an error", "error", err)
		} else {
			b.log.Info("blockscout terminated")
		}

		cmd := exec.Command("docker", "compose", "down")
		cmd.Dir = tempDir
		err := cmd.Run()
		if err != nil {
			b.log.Error("failed to stop blockscout", "error", err)
		}

		// If it stops, signal that the entire app should be closed
		b.closeApp(nil)
		b.stoppedCh <- struct{}{}
	}()

	return nil
}

func (b *Blockscout) dockerComposeEnvs() []string {
	return []string{
		fmt.Sprintf("FRONTEND_PORT=%d", b.config.FrontendPort),
		fmt.Sprintf("BACKEND_PORT=%d", b.config.BackendPort),
		fmt.Sprintf("POSTGRES_PORT=%d", b.config.PostgresPort),
	}
}

func (b *Blockscout) backendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["ETHEREUM_JSONRPC_HTTP_URL"] = b.config.RpcUrl
	envs["ETHEREUM_JSONRPC_TRACE_URL"] = b.config.RpcUrl
	envs["SUBNETWORK"] = b.config.Name
	envs["FIRST_BLOCK"] = fmt.Sprintf("%d", b.config.FirstBlock)
	envs["DATABASE_URL"] = fmt.Sprintf(
		"postgresql://blockscout:ceWb1MeLBEeOIfk65gU8EjF8@host.docker.internal:%v/blockscout", b.config.PostgresPort)
	return envs
}

func (b *Blockscout) frontendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["NEXT_PUBLIC_API_PORT"] = fmt.Sprintf("%d", b.config.BackendPort)
	envs["NEXT_PUBLIC_NETWORK_NAME"] = b.config.Name
	envs["NEXT_PUBLIC_NETWORK_SHORT_NAME"] = b.config.Name
	return envs
}

func (b *Blockscout) registerCleanupTask(task func()) {
	b.cleanupTasks = append(b.cleanupTasks, task)
}

func (b *Blockscout) executeCleanup() {
	for _, task := range b.cleanupTasks {
		task()
	}
}
