package player

import (
	"fmt"
	"os"

	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/internal/ui"
	"github.com/dhilzyi/hianime-cli/providers/hianime"
)

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
