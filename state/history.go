package state

import (
	"github.com/dhilzyi/hianime-cli/cli"

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

var historyFilePath string

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

	if err := os.MkdirAll(cli.PathsData.DataDir, 0755); err != nil {
		return err
	}

	if err = os.WriteFile(historyFilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

func LoadHistory() ([]History, error) {
	var historySession []History

	dataDir := cli.PathsData.DataDir

	historyFilePath = filepath.Join(dataDir, "history.json")
	oldPathHistory := filepath.Join("state", "history.json")

	if data, err := os.ReadFile(historyFilePath); err == nil {
		if err = json.Unmarshal(data, &historySession); err != nil {
			return nil, err
		}

		fmt.Println("Load history success")
		return historySession, err
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if data, err := os.ReadFile(oldPathHistory); err == nil {
		if err = json.Unmarshal(data, &historySession); err != nil {
			return nil, err
		}

		os.MkdirAll(dataDir, 0755)

		if err := SaveHistory(historySession); err != nil {
			return nil, err
		}

		fmt.Println("History migrated from legacy location.")

		return historySession, err
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if err := SaveHistory(historySession); err != nil {
		return nil, err
	}

	return historySession, nil
}
