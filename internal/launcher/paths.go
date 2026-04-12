package launcher

import (
	"github.com/adrg/xdg"
)

func DefaultConfigPath() (string, error) {
	configHome, err := xdg.ConfigFile("eu5-mod-launcher/loadorder.json")
	if err != nil {
		return "", err
	}
	return configHome, nil
}
