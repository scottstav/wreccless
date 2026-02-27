package worker

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// SpawnRun launches "ccl run <id>" as a detached background process.
// The child process survives the parent exiting.
func SpawnRun(id, cclBin, configPath, stateDir string) error {
	cmd := exec.Command(cclBin, "run", id)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Env = append(cleanEnv(),
		"CCL_STATE_DIR="+stateDir,
		"CCL_CONFIG="+configPath,
	)
	return cmd.Start()
}

// cleanEnv returns os.Environ() with CLAUDECODE removed.
// Claude Code sets this var to detect nested sessions — workers must not inherit it.
func cleanEnv() []string {
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "CLAUDECODE=") {
			env = append(env, e)
		}
	}
	return env
}
