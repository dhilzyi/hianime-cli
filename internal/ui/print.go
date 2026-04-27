package ui

import (
	"fmt"
	"os"
	"slices"
	"text/tabwriter"

	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/providers/hianime"
)

func PrintEpisodes(episodes []core.Episode, history state.History) {
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
		if eps.Titles.EnglishTitle != "" {
			title = eps.Titles.EnglishTitle
		} else if eps.Titles.RomajiTitle != "" {
			title = eps.Titles.RomajiTitle
		} else if eps.Titles.KanjiTitle != "" {
			title = eps.Titles.KanjiTitle
		}

		fmt.Fprintf(w, "%s\t[%02d]\t%s\t%s\n", prefix, eps.Number, title, timeInfo)
	}
	w.Flush()

}

func DebugPrint(format string, contents ...any) {
	prefix := "[ DEBUG ] "
	fmt.Println(prefix, contents)
}

func PrintSeries(searchData []hianime.SearchElements, order []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	typeOrders := order

	slices.SortFunc(searchData, func(a, b hianime.SearchElements) int {
		orderA := typeOrder(a.Type, typeOrders)
		orderB := typeOrder(b.Type, typeOrders)

		if orderA != orderB {
			if orderA < orderB {
				return -1
			}
		}
		return 1
	})

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

func PrintServers(servers []core.Server) {
	fmt.Println("\n--- Available Servers ---\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NO.\tNAME\tTYPE")

	for i := range servers {
		inst := servers[i]

		fmt.Fprintf(w, "[%d]\t%s\t%s\n",
			i+1, inst.Name, inst.Type,
		)
	}

	w.Flush()
}

func PrintRecentHistory(history []state.History) {
	fmt.Printf("\n--- Recent History ---\n\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	var title string
	var provider string
	fmt.Fprintln(w, "NO.\tNAME\tTYPE\tPROVIDER")

	for i := range history {
		inst := history[i]
		if inst.EnglishName != "" {
			title = inst.EnglishName
		}
		if inst.JapaneseName != "" {
			title = inst.JapaneseName
		}
		if inst.Provider != "" {
			provider = inst.Provider
		} else {
			provider = "N/A"
		}

		fmt.Fprintf(w, "[%d]\t%s\t%s\t%s\n",
			i+1,
			title,
			"N/A",
			provider,
		)

	}
	w.Flush()
}
