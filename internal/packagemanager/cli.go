package packagemanager

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type CLIArgs []string

func (CLIArgs CLIArgs) String() string {
	return strings.Join(CLIArgs, " ")
}

var runCommand = func(args CLIArgs) (string, string, error) {
	var stdoutBuffer, stderrBuffer bytes.Buffer
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = &stderrBuffer
	err := cmd.Run()
	stdout := stdoutBuffer.String()
	stderr := stderrBuffer.String()
	slog.Info("command execute",
		"command", args,
		"stdout", stdout,
		"stderr", stderr,
		"err", err,
	)
	if err != nil {
		return stdout, stderr, fmt.Errorf("failed to run [%s]. Error [%w]. Stdout: [%s], Stderr: [%s]", args.String(), err, stdout, stderr)
	}
	return stdout, stderr, nil
}
