package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type LaunchService struct{}

var (
	errExecutablePathNotConfigured = errors.New("executable path is not configured")
	errExecutablePathIsDirectory   = errors.New("executable path is a directory")
	errUnsupportedSteamLaunchOS    = errors.New("unsupported os for steam launch")
	errInvalidSteamAppID           = errors.New("invalid steam app id")
)

func NewLaunchService() *LaunchService {
	return &LaunchService{}
}

func (*LaunchService) ValidateExecutable(path string) (string, error) {
	if path == "" {
		return "", errExecutablePathNotConfigured
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve executable path %q: %w", path, err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat executable %q: %w", abs, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%w: %q", errExecutablePathIsDirectory, abs)
	}
	return abs, nil
}

func (*LaunchService) BuildLaunchCommand(exePath string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	applyDetachedProcessAttributes(cmd)
	return cmd
}

func (*LaunchService) BuildSteamLaunchCommand(goos, appID string) (*exec.Cmd, error) {
	normalizedAppID, err := normalizeSteamAppID(appID)
	if err != nil {
		return nil, err
	}
	steamURL := "steam://rungameid/" + normalizedAppID

	switch goos {
	case "windows":
		cmd := exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", steamURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case "darwin":
		cmd := exec.CommandContext(context.Background(), "open", steamURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case "linux":
		cmd := exec.CommandContext(context.Background(), "xdg-open", steamURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedSteamLaunchOS, goos)
	}
}

func normalizeSteamAppID(appID string) (string, error) {
	parsed, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errInvalidSteamAppID, appID)
	}

	return strconv.FormatUint(parsed, 10), nil
}

func (*LaunchService) OpenDirectory(goos, path string) error {
	var cmd *exec.Cmd
	switch goos {
	case "windows":
		cmd = exec.CommandContext(context.Background(), "explorer", path)
	case "darwin":
		cmd = exec.CommandContext(context.Background(), "open", path)
	default:
		cmd = exec.CommandContext(context.Background(), "xdg-open", path)
	}
	return cmd.Start()
}
