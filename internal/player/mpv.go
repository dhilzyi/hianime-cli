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
	"strconv"
	"strings"
	"time"

	"github.com/dhilzyi/hianime-cli/cli"
	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/jimaku"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/internal/ui"
)

//go:embed track.lua
var trackScript string
var ScriptName string = "track.lua"

func BuildDesktopCommands(
	cfg config.Settings,
	metaData core.SeriesData,
	episodeData core.Episode,
	serverData core.Server,
	streamingData core.StreamData,
	historyData state.History,
	datadir string,
	flags cli.FlagsStruct,
) []string {
	// Building title display for mpv
	displayTitle := fmt.Sprintf("[Ep. %d] %s (%s)", episodeData.Number, episodeData.Titles.EnglishTitle, serverData.Name)

	// Making headers
	headerFields := []string{}

	for k, v := range streamingData.Headers {
		headerFields = append(headerFields, fmt.Sprintf("%s: %s", k, v))
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
	// if streamingData.Chapters.End > 0 || streamingData.Outro.Start > 0 {
	// 	chapterPathFile := CreateChapters(streamingData, historyData, episodeData, flags.Debug)
	// 	if chapterPathFile != "" {
	// 		fmt.Println("--> Adding chapters to mpv.")
	// 		args = append(args, fmt.Sprintf("--chapters-file=%s", chapterPathFile))
	// 	}
	// } else {
	// 	fmt.Println("--> Intro & Outro doesn't found. Skip creating chapters.")
	// }

	// Jimaku subtitle command
	if cfg.JimakuEnable {
		jimakuList, err := jimaku.GetSubsJimaku(&metaData, episodeData.Number)
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
