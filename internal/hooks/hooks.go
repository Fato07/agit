package hooks

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/fathindos/agit/internal/config"
)

// Runner executes hook commands configured in the agit config.
type Runner struct {
	hooks   map[string]string
	timeout time.Duration
}

// NewRunner creates a Runner from the config. Returns nil if no hooks are configured.
func NewRunner(cfg *config.Config) *Runner {
	if len(cfg.Hooks) == 0 {
		return nil
	}

	timeout := 30 * time.Second
	if cfg.HookTimeout != "" {
		if d, err := time.ParseDuration(cfg.HookTimeout); err == nil {
			timeout = d
		}
	}

	return &Runner{
		hooks:   cfg.Hooks,
		timeout: timeout,
	}
}

// Fire spawns a goroutine to run the hook command for the given event.
// The env map provides environment variables (AGIT_EVENT, AGIT_REPO, etc.).
// Fire returns immediately; the hook runs asynchronously.
// If no hook is configured for the event, this is a no-op.
func (r *Runner) Fire(event string, env map[string]string) {
	if r == nil {
		return
	}

	command, ok := r.hooks[event]
	if !ok || command == "" {
		return
	}

	// Build environment
	envSlice := make([]string, 0, len(env)+1)
	envSlice = append(envSlice, "AGIT_EVENT="+event)
	for k, v := range env {
		envSlice = append(envSlice, k+"="+v)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Env = append(cmd.Environ(), envSlice...)

		if err := cmd.Run(); err != nil {
			log.Printf("hook %q failed: %v", event, err)
		}
	}()
}
