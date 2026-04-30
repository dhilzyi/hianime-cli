package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var configFilePath string

type Config struct {
	JimakuEnable     bool     `json:"jimaku_enable"`     // for enabling jimaku
	AutoSelectServer bool     `json:"auto_selectserver"` // whether user want use auto select server or manual input server
	MpvPath          string   `json:"mpv_path"`          // manually set mpv path command
	EnglishOnly      bool     `json:"english_only"`      // whether user want importing english subtitle only or not into mpv
	SortType         []string `json:"sort_type"`
	LocalVersion     string   `json:"local_version"`
}

func getDefaultConfig(ver string) Config {
	return Config{
		JimakuEnable:     false,
		AutoSelectServer: true,
		MpvPath:          "",
		EnglishOnly:      true,
		SortType:         []string{"TV", "Movie", "OVA", "Special", "ONA", "Music"},
		LocalVersion:     ver,
	}
}

func LoadConfig(configDir, ver string) (Config, error) {
	defaultConfig := getDefaultConfig(ver)

	var configSession Config
	configFilePath = filepath.Join(configDir, "config.json")
	oldPath := "config.json"

	if configData, err := os.ReadFile(configFilePath); err == nil {
		if err = json.Unmarshal(configData, &configSession); err != nil {
			return Config{}, err
		}

		fmt.Println("Load config success")
		return configSession, nil
	} else if !os.IsNotExist(err) {
		return Config{}, err
	}

	if configData, err := os.ReadFile(oldPath); err == nil {
		if err = json.Unmarshal(configData, &configSession); err != nil {
			return Config{}, err
		}

		if err := os.MkdirAll(configDir, 0755); err != nil {
			return Config{}, err
		}
		if err := saveConfig(configSession); err != nil {
			return Config{}, err
		}

		fmt.Println("Config migrated to new location from legacy location.")

		return configSession, nil
	} else if !os.IsNotExist(err) {
		return Config{}, err
	}

	configSession = defaultConfig

	if err := saveConfig(configSession); err != nil {
		return Config{}, err
	}

	fmt.Println("File config load success.")

	return configSession, nil
}

func saveConfig(rawData Config) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(configFilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil

}

func MigrateConfig(oldCfg Config, ver string) (Config, error) {
	defaultConfig := getDefaultConfig(ver)
	oldCfg.LocalVersion = defaultConfig.LocalVersion

	if err := saveConfig(oldCfg); err != nil {
		return Config{}, err
	}

	return oldCfg, nil
}
