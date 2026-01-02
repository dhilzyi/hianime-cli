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

func getDefaultPath() (string, error) {
	exePath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Failed to get executable path: %w", err)
	}
	defaultPath := filepath.Join(exePath, "state")

	if err = os.MkdirAll(defaultPath, 0755); err != nil {
		return "", fmt.Errorf("Failed to find/create directory: %w", err)
	}

	historyPath := filepath.Join(defaultPath, "history.json")

	return historyPath, nil
}

func UpdateHistory(currentHistory []History, targetData History) []History {
	var cleaned []History

	for i := range currentHistory {
		if currentHistory[i].JapaneseName != targetData.JapaneseName {
			cleaned = append(cleaned, currentHistory[i])
		}
	}

	newHistory := append([]History{targetData}, cleaned...)

	if len(newHistory) > 10 {
		newHistory = newHistory[:10]
	}

	return newHistory
}

func SaveHistory(rawData []History) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return fmt.Errorf("Failed to save the history files: %w", err)
	}

	historyPath, err := getDefaultPath()
	if err != nil {
		return fmt.Errorf("Couldn't find the path: %w", err)
	}

	if err = os.WriteFile(historyPath, jsonData, os.ModePerm); err != nil {
		return fmt.Errorf("Failed to write history files: %w", err)
	}

	return nil
}

func LoadHistory() ([]History, error) {
	var historySession []History

	historyPath, err := getDefaultPath()
	if err != nil {
		return historySession, fmt.Errorf("Couldn't find the path: %w", err)
	}

	if _, err := os.Stat(historyPath); err == nil {
		fmt.Println("File history load success.")
		jsonData, err := os.ReadFile(historyPath)
		if err != nil {
			return historySession, fmt.Errorf("Failed to open json files: %w", err)
		}

		if err = json.Unmarshal(jsonData, &historySession); err != nil {
			return historySession, fmt.Errorf("Failed to convert to struct: %w", err)
		}

	} else if os.IsNotExist(err) {
		_, err := os.Create(historyPath)

		SaveHistory(historySession)

		if err != nil {
			return historySession, fmt.Errorf("Failed to create history json file: %w", err)
		}
	} else {
		return historySession, fmt.Errorf("Error accessing path %s: %w\n", historyPath, err)
	}

	return historySession, nil
}
