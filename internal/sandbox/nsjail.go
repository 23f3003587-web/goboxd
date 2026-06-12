package sandbox

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Nsjail struct {
	// Add config options if your daemon requires parameters later
}

func NewNsjail() *Nsjail {
	return &Nsjail{}
}

// Run cleans up the incoming arguments and executes the isolated sandbox task
func (n *Nsjail) Run(cmd string, args []string) error {
	// 1. Clean the arguments to catch and fix double-slashes or empty variables
	var cleanedArgs []string
	for _, arg := range args {
		// If an argument looks like "//solution.py", clean it to "/app/solution.py"
		// because /tmp/goboxd-jail-xxx is mounted at /app inside the chroot.
		if strings.HasPrefix(arg, "//") {
			arg = "/app/" + strings.TrimPrefix(arg, "//")
		} else if arg == "/solution.py" {
			arg = "/app/solution.py"
		}
		cleanedArgs = append(cleanedArgs, arg)
	}

	// 2. Set up the execution command targeting the nsjail binary inside the container
	// If 'cmd' is already the binary name or path, we pass it right along.
	execCmd := exec.Command(cmd, cleanedArgs...)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// 3. Run the process
	err := execCmd.Run()
	if err != nil {
		return fmt.Errorf("nsjail execution failed: %v (stderr: %s)", err, stderr.String())
	}

	return nil
}
