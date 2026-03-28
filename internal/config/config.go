package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var configFilePath string

type Settings struct {
	JimakuEnable     bool     `json:"jimaku_enable"`     // for enabling jimaku
	AutoSelectServer bool     `json:"auto_selectserver"` // whether user want use auto select server or manual input server
	MpvPath          string   `json:"mpv_path"`          // manually set mpv path command
	EnglishOnly      bool     `json:"english_only"`      // whether user want importing english subtitle only or not into mpv
	SortType         []string `json:"sort_type"`
	LocalVersion     string   `json:"local_version"`
}

func LoadConfig(configDir string) (Settings, error) {
	var configSession Settings
	configFilePath = filepath.Join(configDir, "config.json")
	oldPath := "config.json"

	if configData, err := os.ReadFile(configFilePath); err == nil {
		if err = json.Unmarshal(configData, &configSession); err != nil {
			return Settings{}, err
		}

		fmt.Println("Load config success")
		return Settings{}, nil
	} else if !os.IsNotExist(err) {
		return Settings{}, err
	}

	if configData, err := os.ReadFile(oldPath); err == nil {
		if err = json.Unmarshal(configData, &configSession); err != nil {
			return Settings{}, err
		}

		os.MkdirAll(configDir, 0755)
		if err := SaveConfig(configSession); err != nil {
			return Settings{}, err
		}

		fmt.Println("Config migrated to new location from legacy location.")

		return Settings{}, nil
	} else if !os.IsNotExist(err) {
		return Settings{}, err
	}

	configSession = Settings{
		JimakuEnable:     true,
		AutoSelectServer: true,
		MpvPath:          "",
		EnglishOnly:      true,
		SortType:         []string{"TV", "Movie", "OVA", "Special", "ONA", "Music"},
	}

	if err := SaveConfig(configSession); err != nil {
		return Settings{}, err
	}

	fmt.Println("File config load success.")

	return configSession, nil
}

func SaveConfig(rawData Settings) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(configFilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil

}
