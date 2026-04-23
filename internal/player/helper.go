package player

import (
	"os"
	"runtime"
	"strings"
)

func GetMpvBinary(mpvPath string) string {
	if mpvPath != "" {
		return mpvPath
	}
	if runtime.GOOS == "windows" {
		return "mpv.exe"
	}

	if runtime.GOOS == "linux" {
		if isWSL() {
			return "mpv.exe"
		}
		return "mpv"
	}

	return "mpv"
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, "microsoft") || strings.Contains(content, "wsl")
}
