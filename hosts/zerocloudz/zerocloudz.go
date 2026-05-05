package zerocloudz

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

type rawJson5 struct {
	Type string    `json:"type"`
	Data json5Data `json:"data"`

	RawData map[string]interface{} `json:"-"`
}
type json5Data struct {
	VideoID    string `json:"video_id"`
	VideoTitle string `json:"video_title"`
	Subtitles  []subtitleData
	Chapters   []chapterData

	ObfuscatedCryptoData map[string]interface{} `json:"obfuscated_crypto_data"`
	ObfuscationSeed      string                 `json:"obfuscation_seed"`
}

type subtitleData struct {
	Url      string
	Language string
	Format   string
}

type chapterData struct {
	Start int
	End   int
	Title string
}

type cryptoData struct {
	Key   string // dynamic (kf_xxx)
	IV    string // dynamic (ivf_xxx)
	Token string // dynamic (xxxx_xxxx)
	Extra string // dynamic secondary key
}

type tokenApiResponse struct {
	VideoB64 string `json:"video_b64"`
	KeyFrag  string `json:"key_frag"`
}

func GetStreamData(rawUrl string) (core.StreamData, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return core.StreamData{}, err
	}
	defaultDomain := fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host)
	client, err := common.NewSession()
	if err != nil {
		return core.StreamData{}, err
	}

	rawHtml, err := fetchRawHtml(rawUrl, client)
	if err != nil {
		return core.StreamData{}, err
	}

	extracted, err := extractPayload(rawHtml)
	if err != nil {
		return core.StreamData{}, err
	}
	json5Data, err := parseToJson5(extracted)
	if err != nil {
		return core.StreamData{}, err
	}

	masterUrl, err := getMasterURL(defaultDomain, client, json5Data)
	if err != nil {
		return core.StreamData{}, err
	}

	return core.StreamData{
		Url: masterUrl,
		Headers: map[string]string{
			"Referer": defaultDomain,
		},

		Chapters: buildChaptersData(json5Data.Data.Chapters),
		Tracks:   buildSubtitles(json5Data.Data.Subtitles),
	}, nil

}

func buildChaptersData(data []chapterData) []core.Timestamp {
	var newChapter []core.Timestamp
	for _, chapter := range data {
		newChapter = append(newChapter, core.Timestamp{
			Start: chapter.Start,
			End:   chapter.End,
			Name:  chapter.Title,
		})
	}
	return newChapter
}

func buildSubtitles(data []subtitleData) []core.Track {
	var tracks []core.Track
	for _, track := range data {
		tracks = append(tracks, core.Track{
			Url:      track.Url,
			Type:     track.Format,
			Language: track.Language,
		})
	}

	return tracks
}
func getMasterURL(defaultDomain string, client *http.Client, rawJson rawJson5) (string, error) {
	cryptoData, err := buildCryptoData(rawJson)
	if err != nil {
		return "", err
	}

	tokenApiResp, err := fetchTokenApi(cryptoData.Token, defaultDomain, client)
	if err != nil {
		return "", err
	}

	encryptedKeyBytes, err := cleanBase64(cryptoData.Key)
	if err != nil {
		return "", err
	}

	secondaryKeyBytes, err := cleanBase64(cryptoData.Extra)
	if err != nil {
		return "", err
	}

	dynamicKeyBytes, err := cleanBase64(tokenApiResp.KeyFrag)
	if err != nil {
		return "", err
	}

	ivBytes, err := cleanBase64(cryptoData.IV)
	if err != nil {
		return "", err
	}
	ciphertextBytes, err := cleanBase64(tokenApiResp.VideoB64)
	if err != nil {
		return "", err
	}

	videoUrl, err := performDecryption(rawJson.Data.ObfuscationSeed, encryptedKeyBytes, secondaryKeyBytes, dynamicKeyBytes, ivBytes, ciphertextBytes)
	if err != nil {
		return "", err
	}

	return videoUrl, nil
}
