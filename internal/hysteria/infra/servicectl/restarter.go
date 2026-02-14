package servicectl

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

type ExecCommandRunner struct{}

func (r ExecCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w (%s)", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

type Restarter struct {
	enabled       bool
	serviceName   string
	overrideCmd   string
	commandRunner CommandRunner
}

func NewRestarter(enabled bool, serviceName, overrideCmd string) *Restarter {
	return &Restarter{
		enabled:       enabled,
		serviceName:   serviceName,
		overrideCmd:   strings.TrimSpace(overrideCmd),
		commandRunner: ExecCommandRunner{},
	}
}

func (r *Restarter) Restart(ctx context.Context) error {
	if !r.enabled {
		return nil
	}

	if r.overrideCmd != "" {
		return r.commandRunner.Run(ctx, "sh", "-lc", r.overrideCmd)
	}

	if _, err := exec.LookPath("systemctl"); err == nil {
		return r.commandRunner.Run(ctx, "systemctl", "restart", r.serviceName)
	}

	if _, err := exec.LookPath("service"); err == nil {
		return r.commandRunner.Run(ctx, "service", r.serviceName, "restart")
	}

	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			return r.commandRunner.Run(ctx, "brew", "services", "restart", r.serviceName)
		}
	}

	return fmt.Errorf("no supported service manager found for restarting %q", r.serviceName)
}
