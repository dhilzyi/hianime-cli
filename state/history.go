package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type History struct {
	Url          string                  `json:"url"`
	JapaneseName string                  `json:"jp_name"`
	EnglishName  string                  `json:"en_name"`
	LastEpisode  int                     `json:"last_episode"`
	AnilistID    string                  `json:"anilist_id"`
	SubDelay     float64                 `json:"sub_delay"`
	Volume       int                     `json:"volume"`
	Episode      map[int]EpisodeProgress `json:"episode_history"`
}

type EpisodeProgress struct {
	Position float64 `json:"position"`
	Duration float64 `json:"duration"`
}

var defaultPath string

func UpdateHistory(currentHistory []History, targetData History) []History {
	var cleaned []History

	for i := range currentHistory {
		if currentHistory[i].JapaneseName != targetData.JapaneseName {
			cleaned = append(cleaned, currentHistory[i])
		}
	}

	newHistory := append([]History{targetData}, cleaned...)

	if len(newHistory) > 20 {
		newHistory = newHistory[:20]
	}

	return newHistory
}

func SaveHistory(rawData []History) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(defaultPath, jsonData, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func LoadHistory() ([]History, error) {
	var historySession []History

	if defaultPath == "" {
		var err error
		defaultPath, err = getExePath("history.json")
		if err != nil {
			return []History{}, err
		}
	}

	historyPath := defaultPath

	if _, err := os.Stat(historyPath); err == nil {
		fmt.Println("File history load success.")
		jsonData, err := os.ReadFile(historyPath)
		if err != nil {
			return historySession, err
		}

		if err = json.Unmarshal(jsonData, &historySession); err != nil {
			return historySession, err
		}

	} else if os.IsNotExist(err) {
		_, err := os.Create(historyPath)

		SaveHistory(historySession)

		if err != nil {
			return historySession, err
		}
	} else {
		return historySession, err
	}

	return historySession, nil
}

func getExePath(filename string) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	defaultPath := filepath.Join(filepath.Dir(ex), "state")

	if err = os.MkdirAll(defaultPath, 0755); err != nil {
		return "", err
	}

	historyPath := filepath.Join(defaultPath, filename)

	return historyPath, nil
}
