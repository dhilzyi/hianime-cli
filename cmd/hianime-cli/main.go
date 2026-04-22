package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/cli"
	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/path"
	"github.com/dhilzyi/hianime-cli/internal/player"
	"github.com/dhilzyi/hianime-cli/internal/state"
	"github.com/dhilzyi/hianime-cli/internal/ui"
	"github.com/dhilzyi/hianime-cli/internal/version"
	"github.com/dhilzyi/hianime-cli/providers/hianime"
)

var cacheEpisodes = make(map[string][]hianime.Episodes) // "AnimeID" : {{Eps: 1, ...}, ...}
type InputType int

const (
	InputURL InputType = iota
	InputHistoryIndex
	InputCommand
)

//go:embed version.txt
var embedVersion string

func main() {
	cleanEmbedVersion := strings.TrimSpace(embedVersion)
	flags := cli.ParseFlags()
	cli.HandleFlags(flags, cleanEmbedVersion)

	scanner := bufio.NewScanner(os.Stdin)

	appDir, err := path.InitPath()
	if err != nil {
		fmt.Println("Fail to initialize path: " + err.Error())
	}
	configSes, err := config.LoadConfig(appDir.ConfigDir, cleanEmbedVersion)
	if err != nil {
		fmt.Println("Fail to load config file: " + err.Error())
	}

	if newCfg, updated, err := version.Run(cleanEmbedVersion, appDir.DataDir, configSes); err != nil {
		log.Println(err)
	} else {
		if updated {
			configSes = newCfg
			fmt.Printf("Set to new config complete.\n")
		}
	}

	history, err := state.LoadHistory(appDir.DataDir)
	if err != nil {
		fmt.Println("Fail to load history file: " + err.Error())
	}

seriesLoop:
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
			break seriesLoop
		} else if seriesInput == "s" {
			var searchData []hianime.SearchElements
			var err error

		searchLoop:
			for {
				for {
					fmt.Printf("\nEnter anime name to search (or 'q' to go back):")
					scanner.Scan()
					searchInput := scanner.Text()
					if searchInput == "q" {
						break searchLoop
					}
					searchData, err = hianime.Search(searchInput)
					if err != nil {
						log.Println(err)
						continue
					}

					if len(searchData) != 0 {
						ui.PrintSeries(searchData, configSes.SortType)
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
					break searchLoop
				}
			}

		}

		var historySelect state.History
		var seriesMetadata core.SeriesData
		var provider core.Provider
		var episodes []core.Episode

		inputType, index := classifyInput(seriesInput)
		switch inputType {
		case InputURL:
			provider, episodes, historySelect, seriesMetadata, err = handleURLInput(seriesInput, history)

		case InputHistoryIndex:
			provider, episodes, historySelect, seriesMetadata, err = handleHistoryInput(index, history)

		case InputCommand:
			continue
		}

		history = state.UpdateHistory(history, historySelect)
		if err := state.SaveHistory(history, appDir.DataDir); err != nil {
			log.Println(err)
		}

	episodeLoop:
		for {
			fmt.Printf("\n--- Series: %s ---\n\n", seriesMetadata.Titles.EnglishTitle)

			if episodes != nil {
			}
			// episodes, exists := cacheEpisodes[seriesMetadata.SeriesUrl]
			// if !exists {
			// 	episodes, err = hianime.GetEpisodes(seriesMetadata.AnimeID)
			// 	if err != nil {
			// 		log.Println(err)
			// 		continue
			// 	}
			//
			// 	cacheEpisodes[seriesMetadata.AnimeID] = episodes
			// }

			ui.PrintEpisodes(episodes, historySelect)

			fmt.Print("\nEnter number episode to watch (or 'q' to go back): ")
			scanner.Scan()

			episodeInput := scanner.Text()
			episodeInput = strings.TrimSpace(episodeInput)

			if episodeInput == "q" {
				break episodeLoop
			}

			var selectedNum int
			if episodeInput == "" {
				selectedNum = historySelect.LastEpisode

			} else {
				selectedNum, err = strconv.Atoi(episodeInput)
				if err != nil {
					fmt.Printf("Invaild number: %s", err.Error())
					continue
				}
			}

			var servers []core.Server
			var selectedEpisode core.Episode

			if selectedNum > 0 && selectedNum <= len(episodes) {
				selectedEpisode = episodes[selectedNum-1]
				servers, err = provider.GetServers(selectedEpisode)
				if err != nil {
					log.Println(err)
					continue
				}

				historySelect.LastEpisode = selectedNum

				history = state.UpdateHistory(history, historySelect)
				if err := state.SaveHistory(history, appDir.DataDir); err != nil {
					log.Println(err)
				}
			} else {
				fmt.Println("Error: Number is invalid.")
				continue
			}

			var testedServer int
		serverLoop:
			for {
				if len(servers) == 0 {
					fmt.Println("\nNo available servers found.")
					break
				}

				var selectedServer core.Server
				var streamData core.StreamData

				if configSes.AutoSelectServer {
					if testedServer >= len(servers) {
						fmt.Println("\nNo available servers found for following episode.")
						break
					}

					fmt.Println("\n--> Auto-select server enabled.")

					for i := testedServer; i < len(servers); i++ {
						selectedServer = servers[i]

						fmt.Printf("--> Selecting '%s'....\n", selectedServer.Name)

						attempt, err := provider.GetStreamData(selectedServer.Name)
						if err == nil {
							streamData = attempt
							testedServer = i + 1
							break
						} else {
							log.Println(err)
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
						break serverLoop
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
						fmt.Println("Error: Number is invalid.")
						continue
					}
				}

				if streamData.Url == "" {
					fmt.Println("Error: Couldn't find streamdata url!")
					continue
				}

				// get mpv path automatically according user platforms.
				binName := player.GetMpvBinary(configSes.MpvPath)
				desktopCommands := player.BuildDesktopCommands(configSes, seriesMetadata, selectedEpisode, selectedServer, streamData, historySelect, appDir.DataDir, flags)

				success, subDelay, lastPos, totalDur := player.PlayMpv(binName, desktopCommands, flags.MpvVerbose)

				if success {
					historySelect.SubDelay = subDelay

					if historySelect.Episode == nil {
						historySelect.Episode = make(map[int]state.EpisodeProgress)
					}

					historySelect.Episode[selectedEpisode.Number] = state.EpisodeProgress{
						Position: lastPos,
						Duration: totalDur,
					}

					history = state.UpdateHistory(history, historySelect)
					if err := state.SaveHistory(history, appDir.DataDir); err != nil {
						log.Println(err)
					}

					break serverLoop
				} else {
					continue
				}
			}
		}
	}
}
