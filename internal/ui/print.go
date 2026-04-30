package ui

import (
	"fmt"
	"os"
	"slices"

	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func PrintEpisodes(episodes []core.Episode, history state.History) {
	symbols := tw.NewSymbolCustom("MyGrid").
		WithRow("─").
		WithColumn("│").
		WithCenter("┼")

	table := tablewriter.NewTable(os.Stdout,

		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Borders: tw.BorderNone, // This completely removes the outer box!
			Symbols: symbols,
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenColumns: tw.On, // Turns on vertical lines
				},
				Lines: tw.Lines{
					ShowTop: tw.On, // Keeps the dividing line under the header
				},
			},
		})),

		// Configure the Alignment
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
			Row: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignLeft},
			},
		}),
	)
	table.Header([]string{"LAST", "NO", "EPS NAME", "DURATION"})

	for _, eps := range episodes {
		prefix := "  "
		if eps.Number == history.LastEpisode {
			prefix = "-->"
		}

		timeInfo := ""
		if prog, ok := history.Episodes[eps.Number]; ok {
			curr := prettyDuration(prog.Position)
			total := prettyDuration(prog.Duration)
			timeInfo = fmt.Sprintf("%s/%s", curr, total)
		}
		table.Append([]string{
			prefix,
			fmt.Sprintf("%d", eps.Number),
			common.GetPreferredTitle(eps.Titles),
			timeInfo,
		})
	}
	table.Render()
}

func DebugPrint(format string, contents ...any) {
	prefix := "[ DEBUG ] "
	fmt.Println(prefix, contents)
}

func PrintSearchResults(searchData []core.SearchResult, order []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NO", "NAME", "TYPE", "DURATION", "NUMBER EPS"})

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

	for i := range searchData {
		ins := searchData[i]
		if ins.NumberEpisodes == 0 {
			ins.NumberEpisodes = 1
		}

		table.Append([]string{
			fmt.Sprintf("[%d]", i+1),
			ins.Type,
			fmt.Sprintf("%dm", ins.Duration),
			fmt.Sprintf("%d", ins.NumberEpisodes),
		})
	}

	table.Render()
}

func PrintServers(servers []core.Server) {
	fmt.Printf("\n--- Available Servers ---\n\n")

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NO", "NAME", "TYPE"})

	for i := range servers {
		inst := servers[i]

		table.Append([]string{
			fmt.Sprintf("[%d]", i+1),
			inst.Name,
			inst.Type,
		})
	}

	table.Render()
}

func PrintRecentHistory(history []state.History) {
	fmt.Printf("\n--- Recent History ---\n\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NO", "NAME", "TYPE", "PROVIDER"})

	var provider string
	for i := range history {
		inst := history[i]
		title := common.GetPreferredTitle(inst.Metadata.Titles)
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
