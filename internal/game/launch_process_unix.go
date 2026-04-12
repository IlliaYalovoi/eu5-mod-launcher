//go:build !windows

package game

import (
	"os/exec"
	"syscall"
)

func applyDetachedProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
