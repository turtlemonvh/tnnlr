package tnnlr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
	Pid        int    `json:"pid"` // not set until after process starts
	cmd        *exec.Cmd
}

func (t *Tunnel) getCommand() string {
	remote := t.Host
	if t.Username != "" {
		remote = fmt.Sprintf("%s@%s", t.Username, remote)
	}
	return fmt.Sprintf(`ssh -v -L %d:localhost:%d %s -N`,
		t.LocalPort,
		t.RemotePort,
		remote,
	)
}

// Check if process running the tunnel is alive
// FIXME: Seems to return true even if process has exited
func (t *Tunnel) IsAlive() bool {
	if !t.PortInUse() {
		return false
	}

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

func (t *Tunnel) LogPath() (string, error) {
	return getRelativePath(filepath.Join(relLog, fmt.Sprintf("%s.log", t.Id)))
}

func (t *Tunnel) PidPath() (string, error) {
	return getRelativePath(filepath.Join(relProc, fmt.Sprintf("%s.pid", t.Id)))
}

func (t *Tunnel) PortInUse() bool {
	// Check if this port is already in use
	// https://stackoverflow.com/questions/40296483/continuously-check-if-tcp-port-is-in-use
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort("", fmt.Sprintf("%d", t.LocalPort)), time.Duration(1*time.Millisecond))
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}

// Run the cmd and set the active process
/*
Writes process information into ~/.tnnl/proc/XXX.pid
Writes log information into ~/.tnnl/log/XXX.log
*/
func (t *Tunnel) Run(sshExec string) error {
	var err error

	// Stop if already running
	if err = t.Stop(); err != nil {
		return err
	}

	if t.PortInUse() {
		return fmt.Errorf("Port %d is already in use", t.LocalPort)
	}

	// Set up logging and launch task
	// FIXME: Allow this to live beyond the life of the process
	logPath, err := t.LogPath()
	if err != nil {
		return err
	}
	logOut, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	cmdParts := strings.Split(t.getCommand(), " ")
	cmd := exec.Command(sshExec, cmdParts[1:]...)
	cmd.Stdout = logOut
	cmd.Stderr = logOut
	err = cmd.Start()
	if err != nil {
		return err
	}
	t.cmd = cmd
	t.Pid = cmd.Process.Pid

	// Write JSON representation of task to pid file
	pidPath, err := t.PidPath()
	if err != nil {
		return err
	}
	tJSON, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(pidPath, tJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Stop the tunnel if already running
func (t *Tunnel) Stop() error {
	if t.cmd != nil {
		if t.cmd.Process != nil {
			return t.cmd.Process.Kill()
		}
	}
	// Clear pid file path
	p, err := t.PidPath()
	if err == nil {
		os.Remove(p)
	}
	return nil
}
