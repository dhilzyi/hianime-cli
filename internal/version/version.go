package version

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/migration"
	"golang.org/x/mod/semver"
)

const versionURL = "https://raw.githubusercontent.com/dhilzyi/hianime-cli/main/version.txt"

func Run(embedVer, dataDir string, cfg config.Settings) (config.Settings, bool, error) {
	latestVer, err := getLatestVersion()
	if err != nil {
		return config.Settings{}, false, err
	}

	var newCfg config.Settings
	var updated bool

	if !semver.IsValid(embedVer) || !semver.IsValid(latestVer) {
		return config.Settings{}, false, fmt.Errorf("invalid semver format")
	}

	needMigrate := !semver.IsValid(cfg.LocalVersion) ||
		semver.Compare(cfg.LocalVersion, embedVer) < 0

	if needMigrate {
		newCfg, err = migration.Run(dataDir, embedVer, cfg)
		if err != nil {
			return config.Settings{}, false, err
		}
		updated = true
	}

	resultDecider := updateDecider(embedVer, latestVer)
	if resultDecider != "" {
		fmt.Println(resultDecider)
	}

	return newCfg, updated, nil
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	latest := strings.TrimSpace(string(body))

	return latest, nil
}

func updateDecider(current, latest string) string {
	cmp := semver.Compare(current, latest)

	if cmp == 0 {
		return ""
	} else if cmp < 0 {
		return fmt.Sprintf(
			"New version is available: %s → %s\nRun 'hianime-cli -update' or `go install github.com/dhilzyi/hianime-cli@latest`\n",
			current, latest,
		)
	} else {
		return "You are ahead of latest version (dev build)"
	}
}

func migrationCheck(local, embed string) bool {
	return semver.Compare(local, embed) < 0
}
