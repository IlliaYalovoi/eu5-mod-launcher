package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type LaunchService struct{}

const (
	darwinOS = "darwin"
	linuxOS  = "linux"
)

var (
	errExecutablePathNotConfigured = errors.New("executable path is not configured")
	errExecutablePathIsDirectory   = errors.New("executable path is a directory")
	errUnsupportedSteamLaunchOS    = errors.New("unsupported os for steam launch")
	errUnsupportedOpenURLOS        = errors.New("unsupported os for open url")
	errInvalidSteamAppID           = errors.New("invalid steam app id")
	errInvalidWorkshopItemID       = errors.New("invalid workshop item id")
	errInvalidOpenURL              = errors.New("invalid open url")
	errUnsupportedURLScheme        = errors.New("unsupported url scheme")
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
	case windowsOS:
		cmd := exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", steamURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case darwinOS:
		cmd := exec.CommandContext(context.Background(), "open", steamURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case linuxOS:
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

func (*LaunchService) BuildWorkshopUnsubscribeURL(itemID string) (string, error) {
	normalizedID, err := normalizeWorkshopItemID(itemID)
	if err != nil {
		return "", err
	}
	return "https://steamcommunity.com/sharedfiles/unsubscribe?id=" + normalizedID, nil
}

func normalizeWorkshopItemID(itemID string) (string, error) {
	parsed, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errInvalidWorkshopItemID, itemID)
	}

	return strconv.FormatUint(parsed, 10), nil
}

func (*LaunchService) OpenDirectory(goos, path string) error {
	var cmd *exec.Cmd
	switch goos {
	case windowsOS:
		cmd = exec.CommandContext(context.Background(), "explorer", path)
	case darwinOS:
		cmd = exec.CommandContext(context.Background(), "open", path)
	default:
		cmd = exec.CommandContext(context.Background(), "xdg-open", path)
	}
	return cmd.Start()
}

func (s *LaunchService) OpenURL(goos, rawURL string) error {
	cmd, err := s.BuildOpenURLCommand(goos, rawURL)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start open url command for %q: %w", rawURL, err)
	}
	return nil
}

func (*LaunchService) BuildOpenURLCommand(goos, rawURL string) (*exec.Cmd, error) {
	normalizedURL, err := normalizeOpenURL(rawURL)
	if err != nil {
		return nil, err
	}

	switch goos {
	case windowsOS:
		cmd := exec.CommandContext(context.Background(), "rundll32", "url.dll,FileProtocolHandler", normalizedURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case darwinOS:
		cmd := exec.CommandContext(context.Background(), "open", normalizedURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case linuxOS:
		cmd := exec.CommandContext(context.Background(), "xdg-open", normalizedURL)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	default:
		return nil, fmt.Errorf("%w: %q", errUnsupportedOpenURLOS, goos)
	}
}

func normalizeOpenURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errInvalidOpenURL, rawURL)
	}

	scheme := parsed.Scheme
	if scheme != "http" && scheme != "https" && scheme != "steam" {
		return "", fmt.Errorf("%w: %q", errUnsupportedURLScheme, scheme)
	}
	if parsed.Host == "" && scheme != "steam" {
		return "", fmt.Errorf("%w: %q", errInvalidOpenURL, rawURL)
	}

	return parsed.String(), nil
}
