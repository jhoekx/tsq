package tsq

import (
	"bytes"
	"os/exec"
)

type CommandTask struct {
	Cmd  string
	Args []string
}

func NewCommandTask(Cmd string, Args ...string) CommandTask {
	return CommandTask{
		Cmd:Cmd,
		Args:Args,
	}
}

func (t *CommandTask) Run(arguments interface{}) (data interface{}, err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(t.Cmd, t.Args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result := make(map[string]string)
	result["stdout"] = stdout.String()
	result["stderr"] = stderr.String()
	data = result
	return
}
