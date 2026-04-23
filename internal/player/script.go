package player

import (
	"fmt"
	"os"
	"path/filepath"
)

func ensureTrackScript(dataDir string) (string, error) {
	scriptDir := filepath.Join(dataDir, "scripts")

	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	scriptPath := filepath.Join(scriptDir, ScriptName)
	if _, err := os.Stat(scriptPath); err == nil {
		fmt.Println("--> Lua script exist")
	} else if os.IsNotExist(err) {
		if err := WriteLuaScript(scriptPath); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("error accessing path %s: %w", ScriptName, err)
	}

	return scriptPath, nil
}

func WriteLuaScript(scriptPath string) error {
	err := os.WriteFile(scriptPath, []byte(trackScript), 0644)
	if err != nil {
		return fmt.Errorf("failed to write script :%w", err)
	}
	return nil
}

func TrackScriptPath(dataDir, fileName string) string {
	return filepath.Join(dataDir, "scripts", fileName)
}
