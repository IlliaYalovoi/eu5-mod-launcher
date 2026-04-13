//go:build windows

package main

import (
	"context"
	"os/exec"
	"testing"
)

func TestApplyDetachedProcessAttributes_WindowsFlags(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), "cmd.exe")
	applyDetachedProcessAttributes(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatalf("SysProcAttr = nil, want windows detached attributes")
	}
	flags := cmd.SysProcAttr.CreationFlags
	if flags&windowsCreateNewProcessGroup == 0 {
		t.Fatalf("CreationFlags missing CREATE_NEW_PROCESS_GROUP: %#x", flags)
	}
	if flags&windowsDetachedProcess == 0 {
		t.Fatalf("CreationFlags missing DETACHED_PROCESS: %#x", flags)
	}
}
