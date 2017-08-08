package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Tunnels
type Tunnel struct {
	Id         string `json:"id"`
	Name       string `form:"name" json:"name" binding:"required"`
	DefaultUrl string `form:"defaultUrl" json:"defaultUrl" binding:"required"`
	Host       string `form:"host" json:"host" binding:"required"`
	Username   string `json:"userName"` // can be ""
	LocalPort  int32  `form:"localPort" json:"localPort" binding:"required"`
	RemotePort int32  `form:"remotePort" json:"remotePort" binding:"required"`
	proc       *os.Process
}

func (t *Tunnel) getCommand() string {
	remote := t.Host
	if t.Username != "" {
		remote = fmt.Sprintf("%s@%s", t.Username, remote)
	}
	return fmt.Sprintf(`ssh -L %d:localhost:%d -f %s -N`,
		t.LocalPort,
		t.RemotePort,
		remote,
	)
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
	t.proc = cmd.Process
	return nil
}

// Stop the tunnel if already running
func (t *Tunnel) Stop() error {
	if t.proc != nil {
		return t.proc.Kill()
	}
	return nil
}
