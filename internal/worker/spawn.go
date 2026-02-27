package worker

import (
	"os"
	"os/exec"
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
	cmd.Env = append(os.Environ(),
		"CCL_STATE_DIR="+stateDir,
		"CCL_CONFIG="+configPath,
	)
	return cmd.Start()
}

