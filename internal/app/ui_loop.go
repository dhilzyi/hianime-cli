package app

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/player"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/internal/ui"
)

func (a *App) handleMenu() {
	var selectedHistory *state.History
	var fetchResult *FetchResult
	for {
		if len(a.History) > 0 {
			ui.PrintRecentHistory(a.History)
		} else {
			fmt.Printf("\n--- No recent history found ---\n\n")
		}
		fmt.Print("\nEnter number or paste supported url to play (or 's' for enter search): ")
		a.Scanner.Scan()

		seriesInput := a.Scanner.Text()
		if seriesInput == "q" {
			return
		} else if seriesInput == "s" {
			urlFromSearch := a.handleSearchView()
			if urlFromSearch == "" {
				continue
			} else {
				seriesInput = urlFromSearch
			}
		}
		var err error
		var url string
		var anilistID int

		inputType, value := classifyInput(seriesInput)
		switch inputType {
		case InputURL:
			url = seriesInput

		case InputHistoryIndex:
			selectedHistory, err = getHistoryByIndex(a.History, value)
			if err != nil {
				log.Println(err)
				continue
			}
			url = selectedHistory.Metadata.SeriesUrl
			anilistID = selectedHistory.Metadata.AnilistID

		case InputAnilistID:
			anilistID = value
		case InputCommand:
			continue
		}

		fetchResult, err = ResolveInput(resolveParams{URL: url, AnilistID: anilistID}, a.Cache)
		if err != nil {
			log.Println(err)
			continue
		}
		provider := fetchResult.Provider
		series := fetchResult.SeriesData

		if selectedHistory == nil {
			selectedHistory, err = findOrCreateHistory(a.History, series, provider.Name())
			if err != nil {
				log.Println(err)
				continue
			}
		} else {
			series = selectedHistory.Metadata
		}
		a.SaveHistory(selectedHistory)

		a.handleEpisodeView(provider, series, selectedHistory, fetchResult.Episodes)
	}
}

func (a *App) handleEpisodeView(
	provider core.Provider,
	series core.SeriesData,
	selectedHistory *state.History,
	episodes []core.Episode,
) {
	var servers []core.Server
	var selectedEpisode core.Episode

	for {
		fmt.Printf("\n--- Series: %s ---\n\n", series.Titles.GetPreferredTitle())
		if len(episodes) < 1 {
			fmt.Println("Error: No episodes data is found")
			break
		}
		ui.PrintEpisodes(episodes, *selectedHistory)

		fmt.Print("\nEnter number episode to watch (or 'q' to go back): ")
		a.Scanner.Scan()

		episodeInput := a.Scanner.Text()
		episodeInput = strings.TrimSpace(episodeInput)

		if episodeInput == "q" {
			return
		}

		var selectedNum int
		var err error
		if episodeInput == "" {
			selectedNum = selectedHistory.LastEpisode

		} else {
			selectedNum, err = strconv.Atoi(episodeInput)
			if err != nil {
				fmt.Printf("Invaild number: %s", err.Error())
				continue
			}
		}

		if selectedNum > 0 && selectedNum <= len(episodes) {
			selectedEpisode = episodes[selectedNum-1]
			servers, err = provider.GetServers(selectedEpisode)
			if err != nil {
				log.Println(err)
				continue
			}

			selectedHistory.LastEpisode = selectedNum
			a.SaveHistory(selectedHistory)
		} else {
			fmt.Println("Error: Number is invalid.")
			continue
		}

		a.handleServerView(servers, selectedEpisode, series, provider, selectedHistory)
	}
}

func (a *App) handleServerView(
	servers []core.Server,
	selectedEpisode core.Episode,
	series core.SeriesData,
	provider core.Provider,
	selectedHistory *state.History,
) {
	var testedServer int
	for {
		if len(servers) == 0 {
			fmt.Println("\nNo available servers found.")
			break
		}

		var selectedServer core.Server
		var streamData core.StreamData

		fmt.Printf("\n--> Episode '%d' selected\n", selectedEpisode.Number)
		if a.Config.AutoSelectServer {
			if testedServer >= len(servers) {
				fmt.Println("\nNo available servers found for following episode.")
				break
			}

			fmt.Println("--> Auto-select server enabled.")
			for i := testedServer; i < len(servers); i++ {
				selectedServer = servers[i]

				fmt.Printf("--> Attempt: %d...\nSelecting '%s'....\n", i+1, selectedServer.Name)

				attempt, err := provider.GetStreamData(selectedServer.Name)
				if err == nil {
					streamData = attempt
					break
				} else {
					fmt.Println("Error:", err)
					testedServer = i + 1
					continue
				}
			}
		} else {
			ui.PrintServers(servers)

			fmt.Print("\nEnter server number (or 'q' to go back): ")
			a.Scanner.Scan()

			serverInput := a.Scanner.Text()
			serverInput = strings.TrimSpace(serverInput)
			if serverInput == "q" {
				return
			}

			serverInputInt, err := strconv.Atoi(serverInput)
			if err != nil {
				fmt.Printf("Error when converting to int: %s\n", err.Error())
				continue
			}
			if serverInputInt > 0 && serverInputInt <= len(servers) {
				selectedServer = servers[serverInputInt-1]

				attempt, err := provider.GetStreamData(selectedServer.Name)
				if err == nil {
					streamData = attempt
				} else {
					log.Println(err)
				}
			} else {
				fmt.Println("Error: number is invalid.")
				continue
			}
		}

		if streamData.Url == "" {
			fmt.Println("Error: could not find streamdata url for this episode")
			break
		}

		// get mpv path automatically according user platforms.
		binName := player.GetMpvBinary(a.Config.MpvPath)
		desktopCommands := player.BuildMpvCommands(
			*a.Config,
			series,
			selectedEpisode,
			selectedServer,
			streamData,
			*selectedHistory,
			a.AppDir.DataDir,
			a.Flags.MpvVerbose,
		)

		success, subDelay, lastPos, totalDur := player.PlayMpv(binName, desktopCommands, a.Flags.MpvVerbose)

		if success {
			selectedHistory.SubDelay = subDelay
			if selectedHistory.Episodes == nil {
				selectedHistory.Episodes = make(map[int]state.EpisodeProgress)
			}
			selectedHistory.Episodes[selectedEpisode.Number] = state.EpisodeProgress{
				Position: lastPos,
				Duration: totalDur,
			}

			a.SaveHistory(selectedHistory)
			return
		} else {
			continue
		}
	}
}

func (a *App) handleSearchView() string {
	var provider core.Provider
	for {
		fmt.Println("1. Kuudere")
		fmt.Println("2. AnimeNoSub")
		fmt.Print("\nSelect provider site:")
		a.Scanner.Scan()
		searchInput := a.Scanner.Text()
		searchInput = strings.TrimSpace(searchInput)
		if searchInput == "q" {
			return ""
		}
		if len(searchInput) == 0 || searchInput == "" {
			fmt.Println("Error: invalid input")
			continue
		}

		if searchInput == "1" {
			provider, _ = getProvider(resolveParams{URL: "kuudere"})
		} else {
			provider, _ = getProvider(resolveParams{URL: "animenosub"})
		}
		break
	}
	for {
		fmt.Print("\nInput anime name to search:")
		a.Scanner.Scan()
		searchInput := a.Scanner.Text()
		searchInput = strings.TrimSpace(searchInput)
		if searchInput == "q" {
			return ""
		}
		if len(searchInput) == 0 || searchInput == "" {
			fmt.Println("Error: invalid input")
			continue
		}
		results, err := provider.GetSearchResults(searchInput)
		if err != nil {
			fmt.Println("Error: Fail to retrieve results:", err)
		}
		ui.PrintSearchResults(results, a.Config.SortType)

		fmt.Print("\nInput number anime to watch:")
		a.Scanner.Scan()
		if a.Scanner.Text() == "q" {
			return ""
		}

		searchInt, err := strconv.Atoi(a.Scanner.Text())
		if err != nil {
			fmt.Printf("Error: failed converting to int: %s\n", err.Error())
			continue
		}
		if searchInt > 0 && searchInt <= len(results) {
			return results[searchInt-1].Url
		} else {
			fmt.Println("Error: number is invalid.")
			continue
		}

	}
}
