package service

import "testing"

func TestPlaysetServiceResolveLauncherIndex(t *testing.T) {
	svc := NewPlaysetService(nil)
	preferred := 2
	if got := svc.ResolveLauncherIndex(5, 1, &preferred); got != 2 {
		t.Fatalf("ResolveLauncherIndex() = %d, want 2", got)
	}
	if got := svc.ResolveLauncherIndex(5, 1, nil); got != 1 {
		t.Fatalf("ResolveLauncherIndex() = %d, want 1", got)
	}
	if got := svc.ResolveLauncherIndex(0, -1, nil); got != -1 {
		t.Fatalf("ResolveLauncherIndex() = %d, want -1", got)
	}
}

func TestPlaysetServiceValidateIndex(t *testing.T) {
	svc := NewPlaysetService(nil)
	if err := svc.ValidateIndex(0, 1); err != nil {
		t.Fatalf("ValidateIndex() error = %v", err)
	}
	if err := svc.ValidateIndex(-1, 1); err == nil {
		t.Fatalf("ValidateIndex() error = nil, want error")
	}
}
