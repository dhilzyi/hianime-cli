package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/migration"
	"golang.org/x/mod/semver"
)

const (
	repoOwner = "dhilzyi"
	repoName  = "hianime-cli"

	intervalCheck = 43200
)

func Run(embedVer, dataDir string, cfg *config.Config) (bool, error) {
	latestVer, err := getLatestVersion(cfg)
	if err != nil {
		return false, err
	}

	var updated bool
	if !semver.IsValid(embedVer) || !semver.IsValid(latestVer) {
		return false, fmt.Errorf("invalid semver format")
	}

	needMigrate := !semver.IsValid(cfg.LocalVersion) ||
		semver.Compare(cfg.LocalVersion, embedVer) < 0

	if needMigrate {
		err := migration.Run(dataDir, embedVer, cfg)
		if err != nil {
			return false, err
		}
		updated = true
	}

	resultDecider := updateDecider(embedVer, latestVer)
	if resultDecider != "" {
		fmt.Println(resultDecider)
	}

	return updated, nil
}

func updateDecider(current, latest string) string {
	cmp := semver.Compare(current, latest)

	if cmp == 0 {
		return ""
	} else if cmp < 0 {
		return fmt.Sprintf(
			"New version is available: %s → %s\nRun 'hianime-cli -update' or `go install github.com/dhilzyi/hianime-cli/cmd/hianime-cli@latest`\n",
			current, latest,
		)
	} else {
		return "You are ahead of latest version (dev build)"
	}
}

func migrationCheck(local, embed string) bool {
	return semver.Compare(local, embed) < 0
}

func getLatestVersion(cfg *config.Config) (string, error) {
	var latestVer string
	var err error

	now := time.Now().Unix()
	cacheValid := (now - cfg.LastUpdateCheck) < intervalCheck

	if cacheValid && cfg.LatestVersion != "" {
		latestVer = cfg.LatestVersion
	} else {
		latestVer, err = fetchLatestVersion()
		if err != nil {
			if cfg.LatestVersion != "" {
				latestVer = cfg.LatestVersion
			} else {
				return "", err
			}
		} else {
			cfg.LatestVersion = latestVer
			cfg.LastUpdateCheck = now

			if err := config.SaveConfig(*cfg); err != nil {
				fmt.Println("Warning: Failed to save last update check to config")
			}
		}
	}

	return latestVer, nil
}

func fetchLatestVersion() (string, error) {
	goproxyDefault := "https://proxy.golang.org"
	goproxy := goproxyDefault
	cmd := exec.Command("go", "env", "GOPROXY")
	output, err := cmd.Output()
	if err == nil {
		goproxy = strings.TrimSpace(string(output))
	}

	proxies := strings.Split(goproxy, ",")
	if !slices.Contains(proxies, goproxyDefault) {
		proxies = append(proxies, goproxyDefault)
	}

	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		proxy = strings.TrimRight(proxy, "/")
		if proxy == "direct" || proxy == "off" {
			continue
		}

		url := fmt.Sprintf("%s/github.com/%s/%s/@latest", proxy, repoOwner, repoName)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var version struct{ Version string }
		if err = json.Unmarshal(body, &version); err != nil {
			continue
		}

		return version.Version, nil
	}

	return "", fmt.Errorf("failed to fetch latest version")
}
