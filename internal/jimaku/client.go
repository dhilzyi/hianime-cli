package jimaku

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/anilist"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

type Search []searchElement
type searchElement struct {
	ID         int64  `json:"id"`
	AnilistID  int64  `json:"anilist_id"`
	RomajiName string `json:"name"`
}

type Files []fileElement
type fileElement struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Size int64  `json:"size"`
}

var jimakuBaseUrl string = "https://jimaku.cc"

// Set your JimakuAPI to environment variable or just put it directly in this variable as a string.
var jimakuApi string = os.Getenv("JIMAKU_API_KEY") // or "xxxxxxxxx"

func downloadFile(url string, filePath string) (string, error) {
	cleanPath := strings.TrimRight(filePath, ".")
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
		return "", err
	}

	out, err := os.Create(cleanPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return cleanPath, nil
}

func getFiles(entry_id, episodeNum int) (Files, error) {
	urlFiles := fmt.Sprintf("%s/api/entries/%d/files", jimakuBaseUrl, entry_id)

	req, err := http.NewRequest("GET", urlFiles, nil)
	if err != nil {
		return Files{}, err
	}

	query := req.URL.Query()
	query.Add("episode", fmt.Sprintf("%d", episodeNum))

	req.URL.RawQuery = query.Encode()

	req.Header.Add("Authorization", jimakuApi)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Files{}, err
	}

	defer res.Body.Close()

	var subsFiles Files

	if err = json.NewDecoder(res.Body).Decode(&subsFiles); err != nil {
		return Files{}, err
	}

	// for i := range subsFiles {
	// 	ins := subsFiles[i]
	// 	sizeMB := float64(ins.Size) / (1024 * 1024)
	// 	fmt.Printf("Name: %s\nUrl: %s\nSize: %.2f Mb\n\n", ins.Name, ins.Url, sizeMB)
	// }

	return subsFiles, nil

}

func getIdJimaku(anilistID int) (Search, error) {
	urlSearch := fmt.Sprintf("%s/api/entries/search", jimakuBaseUrl)

	req, err := http.NewRequest("GET", urlSearch, nil)
	if err != nil {
		return Search{}, err
	}
	req.Header.Add("Authorization", jimakuApi)

	query := req.URL.Query()
	query.Add("anime", "true")
	query.Add("anilist_id", fmt.Sprintf("%d", anilistID))
	fmt.Println(query)

	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Search{}, err
	}

	defer res.Body.Close()

	var data Search
	if res.StatusCode != http.StatusOK {
		return Search{}, fmt.Errorf("bad status when querying: %s", res.Status)
	}

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return Search{}, err
	}

	if len(data) == 0 {
		return Search{}, fmt.Errorf("--! Nothing found in Jimaku")
	}

	if data[0].ID <= 0 {
		return Search{}, fmt.Errorf("invalid jimaku id")
	}

	return data, nil
}

func GetSubsJimaku(seriesData *core.SeriesData, episodeNum int) ([]string, error) {
	if jimakuApi == "" {
		return []string{}, fmt.Errorf("no Jimaku API found in the enviroment variable")
	}
	fmt.Println("\n--> jimakuApiKey found. Querying into the Jimaku api....")

	if seriesData.AnilistID == 0 {
		if err := anilist.FillSeriesData(seriesData); err != nil {
			return nil, err
		}
	}

	data, err := getIdJimaku(seriesData.AnilistID)
	if err != nil {
		return []string{}, err
	}

	filesList, err := getFiles(int(data[0].ID), episodeNum)
	if err != nil {
		return []string{}, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	defaultPath := filepath.Join(homeDir, "subtitle")
	re := regexp.MustCompile(`[<>:"/\\|?*\.]`)
	cleanName := re.ReplaceAllString(data[0].RomajiName, "")

	seriesDir := filepath.Join(defaultPath, cleanName)

	if err := os.MkdirAll(seriesDir, 0755); err != nil {
		return []string{}, err
	}

	var nameLists []string
	fmt.Println("--> Files list received. Downloading subtitle....")

	for i := range filesList {
		ins := filesList[i]

		// TODO : Handle zip, 7z, rar formats

		ext := strings.ToLower(path.Ext(ins.Url))
		if ext != ".srt" && ext != ".ass" {
			fmt.Printf("Info: Skipping unsupported format: %s (ext %s)\n", ins.Url, ext)
			continue
		}

		rawFilename := path.Base(ins.Url)

		filename, err := url.QueryUnescape(rawFilename)
		if err != nil {
			filename = rawFilename
		}

		fullPath := filepath.Join(seriesDir, filename)

		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("	File already exists, skip download: %s\n", fullPath)
			nameLists = append(nameLists, fullPath)
			continue
		} else if os.IsNotExist(err) {
			fmt.Printf("	Downloading: %s\n", filename)
		} else {
			fmt.Printf("Error: accessing path %s: %v\n", fullPath, err)
			continue
		}

		downloadedPath, err := downloadFile(ins.Url, fullPath)
		if err != nil {
			fmt.Printf("Failed to download %s (index %d) file", ins.Url, i)
			continue
		}

		nameLists = append(nameLists, downloadedPath)
	}

	if len(nameLists) == 0 {
		return nameLists, fmt.Errorf("--! Failed to retrieve anything")
	}

	return nameLists, nil
}
