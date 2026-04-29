package ui

import (
	"fmt"
	"os"
	"slices"
	"text/tabwriter"

	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/olekukonko/tablewriter"
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

		fmt.Fprintf(w, "%s\t[%02d]\t%s\t%s\n", prefix, eps.Number, common.GetPreferredTitle(eps.Titles), timeInfo)
	}
	w.Flush()

}

func DebugPrint(format string, contents ...any) {
	prefix := "[ DEBUG ] "
	fmt.Println(prefix, contents)
}

func PrintSearchResults(searchData []core.SearchResult, order []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	typeOrders := order

	slices.SortFunc(searchData, func(a, b core.SearchResult) int {
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

		fmt.Fprintf(w, "[%d]\t%s\t%s\t%dm\t%d\n",
			i+1,
			common.GetPreferredTitle(ins.Titles),
			ins.Type,
			ins.Duration,
			ins.NumberEpisodes,
		)
	}

	w.Flush()
}

func PrintServers(servers []core.Server) {
	fmt.Printf("\n--- Available Servers ---\n\n")

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
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NO.", "NAME", "TYPE", "PROVIDER"})

	var title string
	var provider string
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

		table.Append([]string{
			fmt.Sprintf("[%d]", i+1),
			title,
			"N/A",
			provider,
		})
	}
	table.Render()
}
