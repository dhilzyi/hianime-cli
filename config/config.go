package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var FileName string
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
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	FileName = filepath.Join(filepath.Dir(ex), "config.json")

	if _, err := os.Stat(FileName); err == nil {
		jsonData, err := os.ReadFile(FileName)
		if err != nil {
			return err
		}

		if err = json.Unmarshal(jsonData, &ConfigSession); err != nil {
			return err
		}

	} else if os.IsNotExist(err) {
		_, err := os.Create(FileName)

		ConfigSession = Settings{
			JimakuEnable:     true,
			AutoSelectServer: true,
			MpvPath:          "",
			EnglishOnly:      true,
			SortType:         []string{"TV", "Movie", "OVA", "Special", "ONA", "Music"},
		}

		SaveConfig(ConfigSession)

		if err != nil {
			return err
		}
	} else {
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

	if err = os.WriteFile(FileName, jsonData, os.ModePerm); err != nil {
		return err
	}

	return nil

}
