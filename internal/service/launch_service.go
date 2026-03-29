package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type LaunchService struct{}

func NewLaunchService() *LaunchService {
	return &LaunchService{}
}

func (s *LaunchService) ValidateExecutable(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("executable path is not configured")
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
		return "", fmt.Errorf("executable path %q is a directory", abs)
	}
	return abs, nil
}

func (s *LaunchService) BuildLaunchCommand(exePath string, args []string) *exec.Cmd {
	cmd := exec.Command(exePath, args...)
	cmd.Dir = filepath.Dir(exePath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	applyDetachedProcessAttributes(cmd)
	return cmd
}

func (s *LaunchService) BuildSteamLaunchCommand(goos, appID string) (*exec.Cmd, error) {
	switch goos {
	case "windows":
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", "steam://rungameid/"+appID)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case "darwin":
		cmd := exec.Command("open", "steam://rungameid/"+appID)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	case "linux":
		cmd := exec.Command("xdg-open", "steam://rungameid/"+appID)
		applyDetachedProcessAttributes(cmd)
		return cmd, nil
	default:
		return nil, fmt.Errorf("unsupported os %q for steam launch", goos)
	}
}

func (s *LaunchService) OpenDirectory(goos, path string) error {
	var cmd *exec.Cmd
	switch goos {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}
