package domain_test

import (
	"errors"
	"eu5-mod-launcher/internal/domain"
	"testing"
)

func TestParseModID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "ok", input: "123456"},
		{name: "empty", input: "", wantErr: domain.ErrEmptyValue},
		{name: "spaces", input: "abc def", wantErr: domain.ErrInvalidID},
		{name: "category mismatch", input: "category:graphics", wantErr: domain.ErrTypeMismatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ParseModID(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ParseModID() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseModID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseCategoryID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "ok", input: "category:graphics"},
		{name: "missing prefix", input: "graphics", wantErr: domain.ErrInvalidID},
		{name: "empty", input: "", wantErr: domain.ErrEmptyValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ParseCategoryID(tt.input)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ParseCategoryID() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseCategoryID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsePlaysetIndex(t *testing.T) {
	if _, err := domain.ParsePlaysetIndex(0); err != nil {
		t.Fatalf("ParsePlaysetIndex(0) error = %v", err)
	}
	if _, err := domain.ParsePlaysetIndex(-1); !errors.Is(err, domain.ErrOutOfRange) {
		t.Fatalf("ParsePlaysetIndex(-1) error = %v, want ErrOutOfRange", err)
	}
}
