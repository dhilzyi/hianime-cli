package player

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dhilzyi/hianime-cli/cli"
	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/jimaku"
	"github.com/dhilzyi/hianime-cli/providers/hianime"
	"github.com/dhilzyi/hianime-cli/ui"
)

//go:embed track.lua
var trackScript string
var ScriptName string = "track.lua"

func BuildDesktopCommands(
	cfg config.Settings,
	metaData hianime.SeriesData,
	episodeData hianime.Episodes,
	serverData hianime.ServerList,
	streamingData hianime.StreamData,
	historyData state.History,
	datadir string,
	flags cli.FlagsStruct,
) []string {
	// Building title display for mpv
	displayTitle := fmt.Sprintf("%s [Ep. %d] %s (%s)", metaData.JapaneseName, episodeData.Number, episodeData.JapaneseTitle, serverData.Name)

	// Making headers
	headerFields := []string{
		fmt.Sprintf("Referer: %s", streamingData.Referer),
		fmt.Sprintf("User-Agent: %s", streamingData.UserAgent),
		fmt.Sprintf("Origin: %s", "https://megacloud.blog"),
	}
	fullHeaders := strings.Join(headerFields, ",")

	// Main commands
	args := []string{
		streamingData.Url,
		"--ytdl-format=bestvideo+bestaudio/best",
		fmt.Sprintf("--http-header-fields=%s", fullHeaders),
		fmt.Sprintf("--title=%s", displayTitle),
		"--script-opts-append=osc-title=${title}",
	}

	// last position if exist in history
	episodeProgress, exist := historyData.Episode[episodeData.Number]
	if exist {
		args = append(args, fmt.Sprintf("--start=%f", episodeProgress.Position))
	}

	// Chapter command
	if streamingData.Intro.End > 0 || streamingData.Outro.Start > 0 {
		chapterPathFile := CreateChapters(streamingData, historyData, episodeData, flags.Debug)
		if chapterPathFile != "" {
			fmt.Println("--> Adding chapters to mpv.")
			args = append(args, fmt.Sprintf("--chapters-file=%s", chapterPathFile))
		}
	} else {
		fmt.Println("--> Intro & Outro doesn't found. Skip creating chapters.")
	}

	// Jimaku subtitle command
	if cfg.JimakuEnable {
		jimakuList, err := jimaku.GetSubsJimaku(metaData, episodeData.Number)
		if err != nil {
			fmt.Printf("Failed to get subs from jimaku: '%s'\n", err)
			fmt.Printf("--> Skipping Jimaku\n")
		} else {
			if len(jimakuList) > 0 {
				for i := range jimakuList {
					args = append(args, fmt.Sprintf("--sub-file=%s", jimakuList[i]))
				}
			}
		}
	} else {
		fmt.Printf("--> Skipping Jimaku\n")
	}

	// Subs from hianime
	for _, track := range streamingData.Tracks {
		if track.Kind == "thumbnails" {
			continue
		}
		if cfg.EnglishOnly && track.Label != "English" {
			continue
		}

		args = append(args, fmt.Sprintf("--sub-file=%s", track.File))
	}

	// Sub delay history command
	if historyData.SubDelay != 0 {
		fmt.Println("--> Adding sub-delay from history...")
		args = append(args, fmt.Sprintf("--sub-delay=%.1f", historyData.SubDelay))
	}

	// track script & debug command
	scriptLua, err := ensureTrackScript(datadir)
	if err == nil {
		args = append(args, "--scripts-append="+scriptLua)
	} else {
		log.Println(err)
	}

	if flags.MpvVerbose {
		args = append(args, "--v")
	}
	return args
}

// NOTE: For intro and outro in mpv so user can know the timestamps and skip easily.
func CreateChapters(data hianime.StreamData, historyData state.History, episodeData hianime.Episodes, debug bool) string {

	f, err := os.CreateTemp("", "hianime_chapters_*.txt")
	if err != nil {
		fmt.Println("Error while creating temporary file: " + err.Error())
		return ""
	}

	contents := ";FFMETADATA1\n"

	writePart := func(start, end int, title string) {
		contents += "[CHAPTER]\n"
		contents += "TIMEBASE=1/1\n"
		contents += fmt.Sprintf("START=%d\n", start)
		contents += fmt.Sprintf("END=%d\n", end)
		contents += fmt.Sprintf("title=%s\n\n", title)
	}

	if data.Intro.Start > 0 || data.Intro.End > 0 {
		if data.Intro.Start == 0 {
			writePart(data.Intro.Start, data.Intro.End, "Intro")
		} else {
			writePart(0, data.Intro.Start, "Part A")
			writePart(data.Intro.Start, data.Intro.End, "Intro")
		}
	}

	if data.Outro.Start > 0 && data.Outro.End > 0 {
		writePart(data.Intro.End, data.Outro.Start, "Part B")
		writePart(data.Outro.Start, data.Outro.End, "Outro")

		// Using exact duration from history if exist
		episodeProgress, exist := historyData.Episode[episodeData.Number]
		if exist {
			if debug {
				ui.DebugPrint("[CHAPTER]", "History duration exist", debug)
			}
			writePart(data.Outro.End, int(episodeProgress.Duration), "Part C")
		} else {
			if debug {
				ui.DebugPrint("[CHAPTER]", "History duration not exist")
			}
			writePart(data.Outro.End, 9999999, "Part C")
		}
	}

	f.WriteString(contents)
	return f.Name()
}

// Now it supports windows and linux automatically, without hardcoding the mpv path. I hope
func PlayMpv(cmdMain string, args []string, verbose bool) (bool, float64, float64, float64) {
	cmdName := cmdMain

	var streamStarted bool
	var subDelay float64
	var lastPos float64
	var totalDuration float64

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cmd := exec.CommandContext(ctx, cmdName, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed to put stdout: " + err.Error())
	}

	fmt.Println("\n--> Executing mpv commands...")

	if err := cmd.Start(); err != nil {
		fmt.Println("Error while running mpv: " + err.Error())
	}

	scanner := bufio.NewScanner(stdout)
	done := make(chan struct{})
	timer := time.AfterFunc(20*time.Second, func() {
		fmt.Println("\n--> MPV is timeout. Killing process...")
		if cmd.Process != nil {
			streamStarted = false
			cmd.Process.Kill()
		}
	})

	var flag bool
	for scanner.Scan() {
		line := scanner.Text()

		if verbose {
			ui.DebugPrint("[MPV]", line)
		}

		if strings.Contains(line, "(+) Video --vid= ") || strings.Contains(line, "h264") {
			timer.Stop()
			if !flag {
				fmt.Println("\nStream is valid. Opening mpv")
				flag = true
			}

			streamStarted = true
		} else if strings.Contains(line, "::STATUS::") {
			parts := strings.Split(line, "::STATUS::")

			if len(parts) > 0 {
				currentStr, totalStr, found := strings.Cut(parts[1], "/")

				if found {
					current, err := strconv.ParseFloat(currentStr, 64)
					if err != nil {
						fmt.Println("Error while converting to float: " + err.Error())
					}
					total, err := strconv.ParseFloat(strings.TrimSpace(totalStr), 64)
					if err != nil {
						fmt.Println("Error while converting to float: " + err.Error())
					}

					lastPos = current
					totalDuration = total
				}
			}

			continue

		} else if strings.Contains(line, "::SUB_DELAY::") {
			parts := strings.Split(line, "::SUB_DELAY::")

			floatDelay, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				fmt.Printf("Error while converting to float: %v\n", err)
			}

			subDelay = floatDelay

			continue
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Println(err)
	}

	close(done)
	timer.Stop()

	return streamStarted, subDelay, lastPos, totalDuration
}

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

func ensureTrackScript(dataDir string) (string, error) {
	scriptDir := filepath.Join(dataDir, "scripts")

	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return "", fmt.Errorf("Failed to create directory: %w", err)
	}

	scriptPath := filepath.Join(scriptDir, ScriptName)
	if _, err := os.Stat(scriptPath); err == nil {
		fmt.Println("--> Lua script exist")
	} else if os.IsNotExist(err) {
		if err := WriteLuaScript(scriptPath); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("Error accessing path %s: %w\n", ScriptName, err)
	}

	return scriptPath, nil
}

func WriteLuaScript(scriptPath string) error {
	err := os.WriteFile(scriptPath, []byte(trackScript), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write script :%w", err)
	}
	return nil
}

func TrackScriptPath(dataDir, fileName string) string {
	return filepath.Join(dataDir, "scripts", fileName)
}
