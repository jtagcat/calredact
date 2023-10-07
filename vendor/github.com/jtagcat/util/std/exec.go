package std

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunWithStdouts(cmd *exec.Cmd, stderrToErr bool) (stdout, stderr string, err error) {
	stdoutBuilder, stderrBuilder := new(strings.Builder), new(strings.Builder)
	cmd.Stdout, cmd.Stderr = stdoutBuilder, stderrBuilder

	err = cmd.Run()
	stdout, stderr = strings.TrimSpace(stdoutBuilder.String()), strings.TrimSpace(stderrBuilder.String())

	if err != nil && stderrToErr {
		err = fmt.Errorf("%s: %s", err.Error(), stderr)
	}

	return
}
