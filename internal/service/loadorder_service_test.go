package service

import (
	"errors"
	"testing"

	"eu5-mod-launcher/internal/domain"
)

func TestLoadOrderServiceValidateAndNormalize(t *testing.T) {
	svc := NewLoadOrderService()
	got, err := svc.ValidateAndNormalize([]string{"mod1", "", "mod1", "mod2"})
	if err != nil {
		t.Fatalf("ValidateAndNormalize() error = %v", err)
	}
	if len(got) != 2 || got[0] != "mod1" || got[1] != "mod2" {
		t.Fatalf("ValidateAndNormalize() = %v, want [mod1 mod2]", got)
	}
}

func TestLoadOrderServiceValidateAndNormalizeRejectsCategoryID(t *testing.T) {
	svc := NewLoadOrderService()
	_, err := svc.ValidateAndNormalize([]string{"category:graphics"})
	if !errors.Is(err, domain.ErrTypeMismatch) {
		t.Fatalf("ValidateAndNormalize() error = %v, want ErrTypeMismatch", err)
	}
}

func TestLoadOrderServiceToggleEnabled(t *testing.T) {
	svc := NewLoadOrderService()
	state, err := svc.ToggleEnabled([]string{"mod1"}, "mod2", true)
	if err != nil {
		t.Fatalf("ToggleEnabled(enable) error = %v", err)
	}
	if len(state) != 2 || state[1] != "mod2" {
		t.Fatalf("ToggleEnabled(enable) = %v", state)
	}

	state, err = svc.ToggleEnabled(state, "mod1", false)
	if err != nil {
		t.Fatalf("ToggleEnabled(disable) error = %v", err)
	}
	if len(state) != 1 || state[0] != "mod2" {
		t.Fatalf("ToggleEnabled(disable) = %v", state)
	}
}
