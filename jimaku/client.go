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

	"hianime-mpv-go/hianime"
)

type Search []SearchElement
type SearchElement struct {
	ID         int64  `json:"id"`
	AnilistID  int64  `json:"anilist_id"`
	RomajiName string `json:"name"`
}

type Files []FileElement
type FileElement struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Size int64  `json:"size"`
}

var UserAgent = ""
var JimakuBaseUrl string = "https://jimaku.cc"

// Set your JimakuAPI to environment variable or just put it directly in this variable as a string.
var JimakuApi string = os.Getenv("JIMAKU_API_KEY") // or "xxxxxxxxx"

func downloadFile(url string, filePath string) (string, error) {
	cleanPath := strings.TrimRight(filePath, ".")
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create dir: %w", err)
	}

	out, err := os.Create(cleanPath)
	if err != nil {
		return "", fmt.Errorf("Failed to create file %s: %w", cleanPath, err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Couldn't fetch the following url: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error while copying the data to the file: %w", err)
	}

	return cleanPath, nil
}

func getFiles(entry_id, episodeNum int) (Files, error) {
	urlFiles := fmt.Sprintf("%s/api/entries/%d/files", JimakuBaseUrl, entry_id)

	req, err := http.NewRequest("GET", urlFiles, nil)
	if err != nil {
		return Files{}, fmt.Errorf("Failed fetching entry id: %w", err)
	}

	query := req.URL.Query()
	query.Add("episode", fmt.Sprintf("%d", episodeNum))

	req.URL.RawQuery = query.Encode()

	fmt.Println(req.URL.RawPath)

	req.Header.Add("Authorization", JimakuApi)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Files{}, fmt.Errorf("Failed to request entry id: %w", err)
	}

	defer res.Body.Close()

	var subsFiles Files

	if err = json.NewDecoder(res.Body).Decode(&subsFiles); err != nil {
		return Files{}, fmt.Errorf("Failed convert subs files to JSON: %w", err)
	}

	// for i := range subsFiles {
	// 	ins := subsFiles[i]
	// 	sizeMB := float64(ins.Size) / (1024 * 1024)
	// 	fmt.Printf("Name: %s\nUrl: %s\nSize: %.2f Mb\n\n", ins.Name, ins.Url, sizeMB)
	// }

	return subsFiles, nil

}

func GetSubsJimaku(seriesData hianime.SeriesData, episodeNum int) ([]string, error) {
	if JimakuApi == "" {
		return []string{}, fmt.Errorf("No Jimaku API found in the enviroment variable.")
	}
	fmt.Println("\n--> JimakuApiKey found. Querying into the Jimaku api....")

	urlSearch := fmt.Sprintf("%s/api/entries/search", JimakuBaseUrl)

	req, err := http.NewRequest("GET", urlSearch, nil)
	if err != nil {
		return []string{}, fmt.Errorf("Failed when parsing url: %w", err)
	}
	req.Header.Add("Authorization", JimakuApi)

	query := req.URL.Query()
	query.Add("anime", "true")

	if seriesData.AnilistID != "" {
		query.Add("anilist_id", seriesData.AnilistID)
	} else {
		fmt.Println("--> AnilistID not found. Processing with query method.")
		query.Add("query", seriesData.JapaneseName)
	}

	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to request query: %w", err)
	}

	defer res.Body.Close()

	var data Search
	if res.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("Bad status when querying: %s", res.Status)
	}

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return []string{}, fmt.Errorf("Failed to decode to JSON: %w", err)
	}

	if len(data) == 0 {
		return []string{}, fmt.Errorf("--! Nothing found in Jimaku.")
	}

	if data[0].ID <= 0 {
		return []string{}, fmt.Errorf("Invalid ID found.")
	}

	filesList, err := getFiles(int(data[0].ID), episodeNum)
	if err != nil {
		return []string{}, fmt.Errorf("Failed when getting files: %w", err)
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
		return []string{}, fmt.Errorf("Failed to create series directory: %w", err)
	}

	var nameLists []string
	fmt.Println("--> Files list received. Downloading subtitle....")

	for i := range filesList {
		ins := filesList[i]

		// TODO : Handle zip, 7z, rar formats

		ext := strings.ToLower(path.Ext(ins.Url))
		if ext != ".srt" && ext != ".ass" {
			fmt.Printf("Skipping unsupported format: %s (extension %s)\n", ins.Url, ext)
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
			fmt.Printf("Error accessing path %s: %v\n", fullPath, err)
		}

		downloadedPath, err := downloadFile(ins.Url, fullPath)
		if err != nil {
			fmt.Printf("Failed to download %s (index %d) file", ins.Url, i)
			continue
		}

		nameLists = append(nameLists, downloadedPath)
	}

	if len(nameLists) == 0 {
		return nameLists, fmt.Errorf("Failed to retrieve.")
	}

	return nameLists, nil
}
