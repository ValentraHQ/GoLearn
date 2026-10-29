//go:build !linux

package main

import (
	"errors"
	"os/exec"
)

// The entrypoint is the sandbox's enforcement point (process groups, signals,
// kill-on-timeout). It only runs inside the Linux runner image, and refuses to
// run elsewhere instead of silently executing code without those guarantees.
// This file exists so the package still builds (and vets) on development
// machines.

func supported() error {
	return errors.New("sandbox-entrypoint is supported on Linux only; run it inside the runner image")
}

func ownProcessGroup(*exec.Cmd) {}

func signalName(*exec.ExitError) (string, bool) { return "", false }
