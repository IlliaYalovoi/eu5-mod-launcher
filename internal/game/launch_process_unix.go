//go:build !windows

package game

import (
	"os/exec"
	"syscall"
)

func ApplyDetachedProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
