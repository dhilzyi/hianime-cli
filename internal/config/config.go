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

	LatestVersion   string `json:"latest_version"`
	LastUpdateCheck int64  `json:"last_update_check"`
}

func getDefaultConfig(ver string) *Config {
	return &Config{
		JimakuEnable:     false,
		AutoSelectServer: true,
		MpvPath:          "",
		EnglishOnly:      true,
		SortType:         []string{"TV", "Movie", "OVA", "Special", "ONA", "Music"},
		LocalVersion:     ver,
	}
}

func LoadConfig(configDir, embedVer string) (*Config, error) {
	cfg := getDefaultConfig(embedVer)

	configFilePath = filepath.Join(configDir, "config.json")
	fileData, err := os.ReadFile(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := SaveConfig(*cfg); err != nil {
			return nil, err
		}
		fmt.Println("Creating new config success")
		return cfg, nil
	}
	if err := json.Unmarshal(fileData, &cfg); err != nil {
		return nil, err
	}
	fmt.Println("Load config success")
	return cfg, nil
}

func SaveConfig(rawData Config) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(configFilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil

}

func BumpConfig(cfg *Config, newVer string) error {
	cfg.LocalVersion = newVer
	if err := SaveConfig(*cfg); err != nil {
		return err
	}

	return nil
}
