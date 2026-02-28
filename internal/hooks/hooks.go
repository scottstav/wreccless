package hooks

import (
	"bytes"
	"log"
	"os/exec"
	"sync"
	"text/template"
)

type Vars struct {
	ID        string
	Task      string
	Dir       string
	Status    string
	SessionID string
}

func render(tmpl string, vars Vars) (string, error) {
	t, err := template.New("hook").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Fire executes hook commands concurrently and waits for all of them to
// start before returning. Each command is a shell string run via sh -c.
// Template variables are expanded before execution. Failures are logged
// but don't propagate.
func Fire(cmds []string, vars Vars) {
	var wg sync.WaitGroup
	for _, cmdTmpl := range cmds {
		expanded, err := render(cmdTmpl, vars)
		if err != nil {
			log.Printf("hook template error: %v", err)
			continue
		}
		wg.Add(1)
		go func(cmd string) {
			defer wg.Done()
			if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
				log.Printf("hook failed: %s: %v", cmd, err)
			}
		}(expanded)
	}
	wg.Wait()
}
