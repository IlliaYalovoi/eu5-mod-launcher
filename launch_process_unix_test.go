//go:build !windows

package main

import (
	"os/exec"
	"testing"
)

func TestApplyDetachedProcessAttributes_UnixSetsid(t *testing.T) {
	cmd := exec.Command("/bin/sh")
	applyDetachedProcessAttributes(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatalf("SysProcAttr = nil, want unix detached attributes")
	}
	if !cmd.SysProcAttr.Setsid {
		t.Fatalf("SysProcAttr.Setsid = false, want true")
	}
}
