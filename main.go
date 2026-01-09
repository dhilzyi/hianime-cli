package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"hianime-mpv-go/config"
	"hianime-mpv-go/hianime"
	"hianime-mpv-go/player"
	"hianime-mpv-go/state"
	"hianime-mpv-go/ui"
)

var cacheEpisodes = make(map[string][]hianime.Episodes) // "AnimeID" : {{Eps: 1, ...}, ...}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var url string
	history, err := state.LoadHistory()
	if err != nil {
		fmt.Println(err)
	}
	configSession, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Fail to load config file: " + err.Error())
	}

	flag.BoolVar(&config.DebugMode, "debug", false, "Enable verbose debug logging")
	flag.Parse()
series_loop:
	for {
		if len(history) > 0 {
			fmt.Printf("\n--- Recent History ---\n\n")
			for i := range history {
				fmt.Printf(" [%d] %s\n", i+1, history[i].JapaneseName)
			}

		} else {
			fmt.Printf("\n--- No recent history found ---\n\n")
		}
		fmt.Print("\nEnter number or paste hianime url to play (or 's' to call api search): ")
		scanner.Scan()

		seriesInput := scanner.Text()
		if seriesInput == "q" {
			break series_loop
		} else if seriesInput == "s" {
			var searchData []hianime.SearchElements
			var err error
		search_loop:
			for {
				for {
					fmt.Printf("\nEnter anime name to search (or 'q' to go back):")
					scanner.Scan()
					searchInput := scanner.Text()
					searchData, err = hianime.Search(searchInput)
					if err != nil {
						fmt.Println(err)
					}

					if len(searchData) != 0 {
						ui.PrintSeries(searchData)
						break
					} else {
						fmt.Println("--! No anime result found")
						continue
					}
				}

				for {
					fmt.Printf("\nEnter number anime to play (or 'q' to go back): ")
					scanner.Scan()

					usrInput := scanner.Text()
					if usrInput == "q" {
						break
					}

					usrInputInt, err := strconv.Atoi(strings.TrimSpace(usrInput))
					if err != nil {
						fmt.Println("Failed to convert to integer. Input number.")
						continue
					} else if usrInputInt > len(searchData) || usrInputInt <= 0 {
						fmt.Println("Input is out of range.")
						continue
					}

					seriesInput = searchData[usrInputInt-1].Url
					break search_loop
				}
			}

		}

		var historySelect state.History
		var seriesMetadata hianime.SeriesData

		if strings.Contains(seriesInput, "hianime.to") {
			url = seriesInput
			seriesMetadata = hianime.GetSeriesData(url)
			newHistory := state.History{
				Url:          seriesMetadata.SeriesUrl,
				JapaneseName: seriesMetadata.JapaneseName,
				EnglishName:  seriesMetadata.EnglishName,
				AnilistID:    seriesMetadata.AnilistID,
				LastEpisode:  1,
				Episode:      make(map[int]state.EpisodeProgress),
			}
			historySelect = newHistory

			history = state.UpdateHistory(history, newHistory)
			state.SaveHistory(history)
		} else {
			if seriesInput == "q" {
				continue
			}

			seriesInputInt, err := strconv.Atoi(seriesInput)
			if err != nil {
				fmt.Println("Invalid input. Enter number or paste url or use search api with `s`")
				continue
			}

			historySelect = history[seriesInputInt-1]
			url = historySelect.Url

			seriesMetadata = hianime.GetSeriesData(url)

			history = state.UpdateHistory(history, historySelect)
			state.SaveHistory(history)
		}

	episode_loop:
		for {
			fmt.Printf("\n--- Series: %s ---\n\n", seriesMetadata.JapaneseName)

			episodeCache, exists := cacheEpisodes[seriesMetadata.AnimeID]
			if !exists {
				episodeCache = hianime.GetEpisodes(seriesMetadata.AnimeID)
				cacheEpisodes[seriesMetadata.AnimeID] = episodeCache
			}

			ui.PrintEpisodes(episodeCache, historySelect)

			fmt.Print("\nEnter number episode to watch (or 'q' to go back): ")
			scanner.Scan()

			episodeInput := scanner.Text()
			episodeInput = strings.TrimSpace(episodeInput)

			if episodeInput == "q" {
				break episode_loop
			}

			var selectedNum int
			var err error
			if episodeInput == "" {
				selectedNum = historySelect.LastEpisode

			} else {
				selectedNum, err = strconv.Atoi(episodeInput)
				if err != nil {
					fmt.Printf("Invaild number: %s", err.Error())
					continue
				}
			}

			var servers []hianime.ServerList

			var selectedEpisode hianime.Episodes
			if selectedNum > 0 && selectedNum <= len(episodeCache) {
				selectedEpisode = episodeCache[selectedNum-1]
				servers = hianime.GetEpisodeServerId(selectedEpisode.Id)

				historySelect.LastEpisode = selectedNum

				history = state.UpdateHistory(history, historySelect)
				state.SaveHistory(history)
			} else {
				fmt.Println("Number is invalid.")
				continue
			}

			var testedServer int
		server_loop:
			for {
				if len(servers) == 0 {
					fmt.Println("\nNo available servers found.")
					break
				}

				var selectedServer hianime.ServerList
				var streamData hianime.StreamData

				if configSession.AutoSelectServer {
					if testedServer >= len(servers) {
						fmt.Println("\nNo available servers found for following episode.")
						break
					}

					fmt.Println("\n--> Auto-select server enabled.")

					for i := testedServer; i < len(servers); i++ {
						selectedServer = servers[i]

						fmt.Printf("--> Selecting '%s'....\n", selectedServer.Name)

						attempt, err := hianime.GetStreamData(selectedServer.DataId)
						if err == nil {
							streamData = attempt
							testedServer = i + 1
							break
						}
					}

				} else {
					fmt.Print("\n--- Available Servers ---\n")

					for i := range len(servers) {
						serverIns := servers[i]

						if serverIns.Type == "dub" {
							fmt.Printf(" [%d] %s (Dub)\n", i+1, serverIns.Name)
						} else {
							fmt.Printf(" [%d] %s\n", i+1, serverIns.Name)
						}
					}
					fmt.Print("\nEnter server number (or 'q' to go back): ")
					scanner.Scan()

					serverInput := scanner.Text()
					serverInput = strings.TrimSpace(serverInput)

					if serverInput == "q" {
						break server_loop
					}
					serverInputInt, err := strconv.Atoi(serverInput)
					if err != nil {
						fmt.Printf("Error when converting to int: %s\n", err.Error())
						continue
					}

					if serverInputInt > 0 && serverInputInt <= len(servers) {
						selectedServer = servers[serverInputInt-1]

						attempt, err := hianime.GetStreamData(selectedServer.DataId)
						if err == nil {
							streamData = attempt
							fmt.Println(streamData)
						}
					} else {
						fmt.Println("Number is invalid.")
						continue
					}
				}

				if streamData.Url == "" {
					fmt.Println("Couldn't find streamdata url!")
					continue
				}

				// get mpv path automatically according user platforms.
				binName := player.GetMpvBinary(configSession.MpvPath)
				desktopCommands := player.BuildDesktopCommands(seriesMetadata, selectedEpisode, selectedServer, streamData, historySelect, configSession)

				success, subDelay, lastPos, totalDur := player.PlayMpv(binName, desktopCommands)

				if success {
					cleanDelay := math.Round(subDelay*10) / 10
					historySelect.SubDelay = cleanDelay

					if historySelect.Episode == nil {
						historySelect.Episode = make(map[int]state.EpisodeProgress)
					}

					historySelect.Episode[selectedEpisode.Number] = state.EpisodeProgress{
						Position: lastPos,
						Duration: totalDur,
					}

					history = state.UpdateHistory(history, historySelect)
					state.SaveHistory(history)

					break server_loop
				} else {
					continue
				}
			}
		}
	}
}
