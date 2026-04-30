package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

type History struct {
	Metadata     core.SeriesData         `json:"metadata"`
	Episodes     map[int]EpisodeProgress `json:"episode_history"`
	ProviderURLs map[string]string       `json:"provider_urls"`

	// LastProvider string `json:"last_provider"`
	Provider    string `json:"provider"`
	LastEpisode int    `json:"last_episode"`

	SubDelay float64 `json:"sub_delay"`
	Volume   int     `json:"volume"`
}

type LegacyHistory struct {
	Url          string                  `json:"url"`
	JapaneseName string                  `json:"jp_name"`
	EnglishName  string                  `json:"en_name"`
	LastEpisode  int                     `json:"last_episode"`
	AnilistID    int                     `json:"anilist_id"`
	SubDelay     float64                 `json:"sub_delay"`
	Volume       int                     `json:"volume"`
	Episode      map[int]EpisodeProgress `json:"episode_history"`
	Provider     string                  `json:"provider"`
}

type EpisodeProgress struct {
	Position float64 `json:"position"`
	Duration float64 `json:"duration"`
}

var historyFilePath string

func UpdateHistory(currentHistory []History, targetData History) []History {
	var cleaned []History

	for i := range currentHistory {
		if currentHistory[i].Metadata.Titles.RomajiTitle != targetData.Metadata.Titles.RomajiTitle {
			cleaned = append(cleaned, currentHistory[i])
		} else if currentHistory[i].Metadata.Titles.EnglishTitle != targetData.Metadata.Titles.EnglishTitle {
			cleaned = append(cleaned, currentHistory[i])
		} else if currentHistory[i].Metadata.AnilistID != targetData.Metadata.AnilistID {
			cleaned = append(cleaned, currentHistory[i])
		}
	}

	newHistory := append([]History{targetData}, cleaned...)

	if len(newHistory) > 20 {
		newHistory = newHistory[:20]
	}

	return newHistory
}

func SaveHistory(rawData []History, dataDir string) error {
	jsonData, err := json.MarshalIndent(rawData, "", " ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	if err = os.WriteFile(historyFilePath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

func LoadHistory(dataDir string) ([]History, error) {
	var historySession []History

	historyFilePath = filepath.Join(dataDir, "history.json")
	if err := MigrateHistoryIfNeeded(historyFilePath); err != nil {
		fmt.Println("Warning: Fail to check migrate history:" + err.Error())
	}

	oldPathHistory := filepath.Join("state", "history.json")

	if data, err := os.ReadFile(historyFilePath); err == nil {
		if err := json.Unmarshal(data, &historySession); err != nil {
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

		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, err
		}

		if err := SaveHistory(historySession, dataDir); err != nil {
			return nil, err
		}

		fmt.Println("History migrated from legacy location.")

		return historySession, err
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if err := SaveHistory(historySession, dataDir); err != nil {
		return nil, err
	}

	return historySession, nil
}

func (h *LegacyHistory) UnmarshalJSON(data []byte) error {
	type Alias LegacyHistory

	aux := &struct {
		AnilistID any `json:"anilist_id"`
		*Alias
	}{
		Alias: (*Alias)(h),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.AnilistID.(type) {
	case float64: // normal number
		h.AnilistID = int(v)
	case string: // old format
		if v == "" {
			h.AnilistID = 0
			return nil
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		h.AnilistID = id
	case nil:
		h.AnilistID = 0
	default:
		return fmt.Errorf("invalid type for anilist_id")
	}

	return nil
}

// MigrateHistoryIfNeeded checks the file and upgrades it to the new format if necessary
func MigrateHistoryIfNeeded(filePath string) error {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var rawCheck []map[string]any
	if err := json.Unmarshal(fileData, &rawCheck); err != nil {
		return err
	}

	if len(rawCheck) == 0 {
		return nil
	}

	if _, hasMetadata := rawCheck[0]["metadata"]; hasMetadata {
		return nil
	}

	fmt.Println("Info: Legacy history detected. Upgrading to new format...")

	var legacyData []LegacyHistory
	if err := json.Unmarshal(fileData, &legacyData); err != nil {
		return err
	}

	var newData []History
	for _, old := range legacyData {

		urls := make(map[string]string)
		if old.Provider != "" && old.Url != "" {
			urls[old.Provider] = old.Url
		}

		migrated := History{
			Metadata: core.SeriesData{
				AnilistID: old.AnilistID,
				Titles: core.Title{
					EnglishTitle: old.EnglishName,
					RomajiTitle:  old.JapaneseName,
				},
				SeriesUrl: old.Url,
			},
			Provider:     old.Provider,
			ProviderURLs: urls,
			LastEpisode:  old.LastEpisode,
			SubDelay:     old.SubDelay,
			Volume:       old.Volume,
			Episodes:     old.Episode,
		}

		newData = append(newData, migrated)
	}

	newBytes, err := json.MarshalIndent(newData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, newBytes, 0644)
}
