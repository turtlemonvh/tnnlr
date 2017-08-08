package tnnlr

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// Tunnels
type Tunnel struct {
	Id         string `json:"id"`
	Name       string `form:"name" json:"name" binding:"required"`
	DefaultUrl string `form:"defaultUrl" json:"defaultUrl" binding:"required"`
	Host       string `form:"host" json:"host" binding:"required"`
	Username   string `form:"username" json:"userName"` // can be ""
	LocalPort  int32  `form:"localPort" json:"localPort" binding:"required"`
	RemotePort int32  `form:"remotePort" json:"remotePort" binding:"required"`
	cmd        *exec.Cmd
}

func (t *Tunnel) getCommand() string {
	remote := t.Host
	if t.Username != "" {
		remote = fmt.Sprintf("%s@%s", t.Username, remote)
	}
	return fmt.Sprintf(`ssh -L %d:localhost:%d %s -N`,
		t.LocalPort,
		t.RemotePort,
		remote,
	)
}

// Check if process running the tunnel is alive
// FIXME: Seems to return true even if process has exited
func (t *Tunnel) IsAlive() bool {
	if t.cmd == nil {
		return false
	}
	if t.cmd.Process == nil {
		return false
	}

	// https://stackoverflow.com/questions/15204162/check-if-a-process-exists-in-go-way
	if err := t.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	// ProcessState is set after a call to `Wait` or `Run`
	if t.cmd.ProcessState != nil {
		return t.cmd.ProcessState.Exited()
	}

	return true
}

// FIXME: Check required fields
func (t *Tunnel) Validate() error {
	return nil
}

// Run the cmd and set the active process
func (t *Tunnel) Run(sshExec string) error {
	var err error

	// Stop if already running
	if err = t.Stop(); err != nil {
		return err
	}

	cmdParts := strings.Split(t.getCommand(), " ")
	cmd := exec.Command(sshExec, cmdParts[1:]...)
	err = cmd.Start()
	if err != nil {
		return err
	}
	t.cmd = cmd
	return nil
}

// Stop the tunnel if already running
func (t *Tunnel) Stop() error {
	if t.cmd != nil {
		if t.cmd.Process != nil {
			return t.cmd.Process.Kill()
		}
	}
	return nil
}
