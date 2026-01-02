package ui

import (
	"fmt"
	"os"
	"text/tabwriter"

	"hianime-mpv-go/config"
	"hianime-mpv-go/hianime"
	"hianime-mpv-go/state"
)

func prettyDuration(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func PrintEpisodes(episodes []hianime.Episodes, history state.History) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\tNO.\tEPS_NAME\tDURATION")

	for _, eps := range episodes {
		prefix := "  "
		if eps.Number == history.LastEpisode {
			prefix = "-->"
		}

		timeInfo := ""
		if prog, ok := history.Episode[eps.Number]; ok {
			curr := prettyDuration(prog.Position)
			total := prettyDuration(prog.Duration)
			timeInfo = fmt.Sprintf("%s/%s", curr, total)
		}

		var title string
		if eps.JapaneseTitle == "" {
			title = eps.EnglishTitle
		} else {
			title = eps.JapaneseTitle
		}
		fmt.Fprintf(w, "%s\t[%02d]\t%s\t%s\n", prefix, eps.Number, title, timeInfo)
	}
	w.Flush()

}

func DebugPrint(format string, contents ...any) {
	if config.DebugMode {
		prefix := "[ DEBUG ] "
		fmt.Println(prefix, contents)
	}
}

func PrintSeries(searchData []hianime.SearchElements) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "NO.\tNAME\tTYPE\tDURATION\tNUMBER EPS")
	for i := range searchData {
		ins := searchData[i]
		if ins.NumberEpisodes == 0 {
			ins.NumberEpisodes = 1
		}

		fmt.Fprintf(w, "[%d]\t%s\t%s\t%s\t%d\n",
			i+1,
			ins.JapaneseName,
			ins.Type,
			ins.Duration,
			ins.NumberEpisodes,
		)
	}

	w.Flush()
}
