package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"log"
	"net/http"
)

var BaseUrl string = "https://hianime.to"

type SeriesData struct {
	Page             string `json:"page"`
	Name             string `json:"name"`
	AnimeID          string `json:"anime_id"`
	MalID            string `json:"mal_id"`
	AnilistID        string `json:"anilist_id"`
	SeriesURL        string `json:"series_url"`
	SelectorPosition string `json:"selector_position"`
}

type Episodes struct {
	Number        int
	EnglishTitle  string
	JapaneseTitle string
	Url           string
	Id            int
}

type ServerList struct {
	Type   string
	Name   string
	DataId int
	Id     int
}
type AjaxResponse struct {
	Status bool   `json:"status"`
	Html   string `json:"html"`
}

func GetEpisodes(series_url string) []Episodes {
	resp, err := http.Get(series_url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	syncData := doc.Find("#syncData")

	rawJson := syncData.Text()
	var data SeriesData
	json.Unmarshal([]byte(rawJson), &data)

	api_url := fmt.Sprintf("%s/ajax/v2/episode/list/%s", BaseUrl, data.AnimeID)

	api_resp, err := http.Get(api_url)
	if err != nil {
		log.Fatal(err)
	}

	defer api_resp.Body.Close()
	var json_resp AjaxResponse
	if err := json.NewDecoder(api_resp.Body).Decode(&json_resp); err != nil {
		panic("Failed to decode JSON: " + err.Error())
	}

	api_doc, err := goquery.NewDocumentFromReader(strings.NewReader(json_resp.Html))
	if err != nil {
		log.Fatal(err)
	}

	var episodes []Episodes

	api_doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			fmt.Println("Couldn't found href.")
		}

		data_id, exists := s.Attr("data-id")
		if !exists {
			fmt.Println("Couldn't found data-id.")
		}

		id_int, err := strconv.Atoi(data_id)
		if err != nil {
			fmt.Print("Failed to convert to integer: " + err.Error())
		}

		titleDiv := s.Find(".ep-name")
		englishTitle := html.UnescapeString(titleDiv.Text())

		japaneseTitle := ""
		rawJName, exists := titleDiv.Attr("data-jname")
		if exists {
			japaneseTitle = html.UnescapeString(rawJName)
		}

		data_structure := Episodes{
			Number:        i + 1,
			EnglishTitle:  englishTitle,
			JapaneseTitle: japaneseTitle,
			Url:           BaseUrl + html.UnescapeString(href),
			Id:            id_int,
		}
		episodes = append(episodes, data_structure)
	})

	return episodes

}

func GetEpisodeServerId(episode_id int) []ServerList {
	servers_url := fmt.Sprintf("%s/ajax/v2/episode/servers?episodeId=%d", BaseUrl, episode_id)

	server_resp, err := http.Get(servers_url)
	if err != nil {
		fmt.Println("Error while requesting server urls: " + err.Error())
	}
	defer server_resp.Body.Close()

	var server_json AjaxResponse
	if err := json.NewDecoder(server_resp.Body).Decode(&server_json); err != nil {
		fmt.Println("Error while converting to json: " + err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(server_json.Html))
	if err != nil {
		fmt.Println("Failed fecthing json to doc: ", err.Error())
	}

	var servers_list []ServerList

	doc.Find(".server-item").Each(func(i int, s *goquery.Selection) {
		data_type, exists := s.Attr("data-type")
		if !exists {
			fmt.Println("Couldn't found data-type: " + err.Error())
		}
		data_id, exists := s.Attr("data-id")
		if !exists {
			fmt.Println("Couldn't found data-id: " + err.Error())
		}
		int_data_id, err := strconv.Atoi(data_id)
		if err != nil {
			fmt.Println("Failed to convert data_id to int: " + err.Error())
		}

		name := s.Find("a").Text()
		instance := ServerList{
			Type:   data_type,
			Name:   name,
			DataId: int_data_id,
		}

		servers_list = append(servers_list, instance)
	})

	return servers_list
}

func GetMegacloud(server_id int) {
	server := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%d", BaseUrl, server_id)

	fmt.Println(server)
}

func main() {
	url := "https://hianime.to/planetes-210"
	scanner := bufio.NewScanner(os.Stdin)

	var cache_episodes []Episodes
episode_loop:
	for {
		if len(cache_episodes) > 0 {
			for i := range len(cache_episodes) {
				eps := cache_episodes[i]
				fmt.Printf(" [%d] %s ID: %d\n", eps.Number, eps.JapaneseTitle, eps.Id)
			}
		} else {
			cache_episodes = GetEpisodes(url)
			for i := range len(cache_episodes) {
				eps := cache_episodes[i]
				fmt.Printf(" [%d] %s ID: %d\n", eps.Number, eps.JapaneseTitle, eps.Id)
			}
		}

		fmt.Print("\nEnter number episode to watch (or q to go back): ")
		scanner.Scan()

		eps_input := scanner.Text()
		eps_input = strings.TrimSpace(eps_input)

		if eps_input == "q" {
			break episode_loop
		}

		int_input, err := strconv.Atoi(eps_input)
		if err != nil {
			fmt.Printf("Error when converting to int: %s\n", err.Error())
			continue
		}

		var servers []ServerList
		if int_input > 0 && int_input < len(cache_episodes) {
			selected := cache_episodes[int_input-1]
			fmt.Printf("Episode : %d \nTitle: %s \nUrl: %s\n\n", selected.Number, selected.JapaneseTitle, selected.Url)
			servers = GetEpisodeServerId(selected.Id)
		} else {
			fmt.Println("Number is invalid.")
		}
	server_loop:
		for {
			for i := range len(servers) {
				ins := servers[i]

				if ins.Type == "dub" {
					fmt.Printf(" [%d] %s (Dub)\n", i+1, ins.Name)
				} else {
					fmt.Printf(" [%d] %s\n", i+1, ins.Name)
				}
			}
			scanner.Scan()

			server_input := scanner.Text()
			server_input = strings.TrimSpace(server_input)
			break server_loop
		}
	}

}
