package player

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"hianime-mpv-go/hianime"
)

// NOTE: For intro and outro in mpv so user can know the timestamps and skip easily.
func CreateChapters(data hianime.StreamData) string {

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

	if data.Intro.Start == 0 {
		writePart(data.Intro.Start, data.Intro.End, "Intro")
	} else {
		writePart(0, data.Intro.Start, "Part A")
		writePart(data.Intro.Start, data.Intro.End, "Intro")
	}

	writePart(data.Intro.End, data.Outro.Start, "Part B")
	writePart(data.Outro.Start, data.Outro.End, "Outro")
	writePart(data.Outro.End, 9999999, "Part C")

	f.WriteString(contents)
	return f.Name()
}

// TODO: Support other platforms.
func PlayMpv(mpv_commands []string) (bool, float64, float64, float64) {
	cmdName := "mpv.exe"

	track_script := "player/track.lua"
	mpv_commands = append(mpv_commands, "--script="+track_script)

	var stream_started bool
	var subDelay float64
	var lastPos float64
	var totalDuration float64

	cmd := exec.Command(cmdName, mpv_commands...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed to put stdout: " + err.Error())
	}

	fmt.Println("\n--> Executing mpv commands...")

	if err := cmd.Start(); err != nil {
		fmt.Println("Error while running mpv: " + err.Error())
	}

	scanner := bufio.NewScanner(stdout)
	timer := time.AfterFunc(20*time.Second, func() {
		fmt.Println("\n--> MPV is timeout. Killing process...")
		cmd.Process.Kill()
		stream_started = false
	})

	var flag bool
	for scanner.Scan() {
		line := scanner.Text()

		// fmt.Println(line)

		if strings.Contains(line, "(+) Video --vid= ") || strings.Contains(line, "h264") {
			timer.Stop()
			if !flag {
				fmt.Println("\nStream is valid. Opening mpv")
				flag = true
			}

			stream_started = true
		} else if strings.Contains(line, "Opening failed") || strings.Contains(line, "HTTP error") {
			fmt.Println("Failed to stream. Potentially dead link...")

			stream_started = false
			timer.Stop()
			cmd.Process.Kill()

			break
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
	}

	return stream_started, subDelay, lastPos, totalDuration
}
