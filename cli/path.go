package cli

import (
	"os"
	"path/filepath"
	"runtime"
)

var appName string = "hianimecli"
var PathsData AppPaths

type AppPaths struct {
	ConfigDir string
	DataDir   string
}

func GetAppPath(appName string) (AppPaths, error) {
	var configDir, dataDir string
	var err error

	switch runtime.GOOS {
	case "windows":
		configDir, err = os.UserConfigDir()
		if err != nil {
			return AppPaths{}, err
		}
		configDir = filepath.Join(configDir, appName)
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData, err = os.UserCacheDir()
			if err != nil {
				return AppPaths{}, err
			}
		}
		dataDir = filepath.Join(localAppData, appName)

	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return AppPaths{}, err
		}

		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config", appName)
		}
		dataDir = os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share", appName)
		}
	}

	return AppPaths{
		ConfigDir: configDir,
		DataDir:   dataDir,
	}, nil
}

func InitPath() error {
	var err error
	PathsData, err = GetAppPath(appName)
	if err != nil {
		return err
	}

	return nil
}
