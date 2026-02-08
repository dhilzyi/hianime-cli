package config

import (
	"github.com/dhilzyi/hianime-cli/cli"

	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var configFilePath string
var DebugMode bool
var ConfigSession Settings

type Settings struct {
	JimakuEnable     bool     `json:"jimaku_enable"`     // for enabling jimaku
	AutoSelectServer bool     `json:"auto_selectserver"` // whether user want use auto select server or manual input server
	MpvPath          string   `json:"mpv_path"`          // manually set mpv path command
	EnglishOnly      bool     `json:"english_only"`      // whether user want importing english subtitle only or not into mpv
	SortType         []string `json:"sort_type"`
}

func LoadConfig() error {
	configDir := cli.PathsData.ConfigDir
	configFilePath = filepath.Join(configDir, "config.json")
	oldPath := "config.json"

	if configData, err := os.ReadFile(configFilePath); err == nil {
		if err = json.Unmarshal(configData, &ConfigSession); err != nil {
			return err
		}

		fmt.Println("Load config success")
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if configData, err := os.ReadFile(oldPath); err == nil {
		if err = json.Unmarshal(configData, &ConfigSession); err != nil {
			return err
		}

		os.MkdirAll(configDir, 0755)
		if err := SaveConfig(ConfigSession); err != nil {
			return err
		}

		fmt.Println("Config migrated to new location from legacy location.")

		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	ConfigSession = Settings{
		JimakuEnable:     true,
		AutoSelectServer: true,
		MpvPath:          "",
		EnglishOnly:      true,
		SortType:         []string{"TV", "Movie", "OVA", "Special", "ONA", "Music"},
	}

	if err := SaveConfig(ConfigSession); err != nil {
		return err
	}

	fmt.Println("File config load success.")

	return nil
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
