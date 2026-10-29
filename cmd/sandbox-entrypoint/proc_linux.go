//go:build linux

package main

import (
	"os/exec"
	"syscall"
)

// supported reports whether this build can enforce the sandbox's process control.
func supported() error { return nil }

// ownProcessGroup starts cmd in its own process group and makes cancellation
// kill the whole group, so a program cannot leave children running after a
// timeout.
func ownProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
}

// signalName describes the signal that killed a process, if any.
func signalName(err *exec.ExitError) (string, bool) {
	if ws, ok := err.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return ws.Signal().String(), true
	}
	return "", false
}
