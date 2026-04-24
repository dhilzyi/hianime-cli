package kuudere

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type RawJson5 struct {
	Type string    `json:"type"`
	Data Json5Data `json:"data"`

	RawData map[string]interface{} `json:"-"`
}
type Json5Data struct {
	VideoID    string `json:"video_id"`
	VideoTitle string `json:"video_title"`
	Subtitles  []SubtitleData
	Chapters   []ChapterData

	ObfuscatedCryptoData map[string]interface{} `json:"obfuscated_crypto_data"`
	ObfuscationSeed      string                 `json:"obfuscation_seed"`
}

type SubtitleData struct {
	Url      string
	Language string
	Format   string
}

type ChapterData struct {
	Start int
	End   int
	Title string
}

type CryptoData struct {
	Key      string   // dynamic (kf_xxx)
	IV       string   // dynamic (ivf_xxx)
	Token    string   // dynamic (xxxx_xxxx)
	Extra    string   // dynamic secondary key
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
	Encoding  string `json:"encoding"`
}

func ExtractPayload(html string) (string, error) {
	pattern := `data:\s*\[null,null,(\{.*?\})\],\s*form:\snull`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 1 {
		return "", fmt.Errorf("could not find SvelteKit data payload")
	}

	jsonPayload := matches[1]

	return jsonPayload, nil
}

func ParseToJson5(raw string) (RawJson5, error) {
	var result RawJson5
	if err := json5.Unmarshal([]byte(raw), &result); err != nil {
		return RawJson5{}, err
	}

	var fullMap map[string]interface{}
	json5.Unmarshal([]byte(raw), &fullMap)
	if dataMap, ok := fullMap["data"].(map[string]interface{}); ok {
		result.RawData = dataMap
	}

	return result, nil
}

func Decrypt(data RawJson5) (*CryptoData, error) {
	primary := sha256Hex(data.Data.ObfuscationSeed)
	secondary := sha256Hex(primary)

	dynamic := map[string]string{
		"video_field":         "vf_" + primary[0:8],
		"key_field":           "kf_" + primary[8:16],
		"iv_field":            "ivf_" + primary[16:24],
		"container_field":     "cd_" + primary[24:32],
		"array_field":         "ad_" + primary[32:40],
		"object_field":        "od_" + primary[40:48],
		"token_field":         primary[48:64] + "_" + primary[56:64],
		"secondary_key_field": secondary[0:16] + "_" + secondary[16:24],
	}

	ocd := data.Data.ObfuscatedCryptoData

	containerRaw, ok := ocd[dynamic["container_field"]]
	if !ok {
		return nil, fmt.Errorf("container field not found")
	}
	container, ok := getMap(containerRaw)
	if !ok {
		return nil, fmt.Errorf("container not a map")
	}

	arrayRaw, ok := container[dynamic["array_field"]]
	if !ok {
		return nil, fmt.Errorf("array field not found")
	}
	array, ok := getSlice(arrayRaw)
	if !ok || len(array) == 0 {
		return nil, fmt.Errorf("array invalid")
	}

	item, ok := getMap(array[0])
	if !ok {
		return nil, fmt.Errorf("array item not a map")
	}

	objectRaw, ok := item[dynamic["object_field"]]
	if !ok {
		return nil, fmt.Errorf("object field not found")
	}
	object, ok := getMap(objectRaw)
	if !ok {
		return nil, fmt.Errorf("object not a map")
	}

	key, _ := getString(object[dynamic["key_field"]])
	iv, _ := getString(object[dynamic["iv_field"]])

	secondaryKey, _ := getString(data.RawData[dynamic["secondary_key_field"]])
	token, _ := getString(data.RawData[dynamic["token_field"]])

	metaMap, _ := getMap(object["metadata"])

	result := &CryptoData{
		Key:   key,
		IV:    iv,
		Token: token,
		Extra: secondaryKey,
		Metadata: Metadata{
			Timestamp: int64(metaMap["timestamp"].(float64)),
			Version:   metaMap["version"].(string),
			Encoding:  metaMap["encoding"].(string),
		},
	}

	return result, nil
}

func fetchTokenApi(tokenReference string) error {
	url := fmt.Sprintf("%s/api/m3u8/%s/", "", tokenReference)
	return nil
}

func sha256Hex(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func getMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}

func getSlice(v interface{}) ([]interface{}, bool) {
	s, ok := v.([]interface{})
	return s, ok
}

func getString(v interface{}) (string, bool) {
	s, ok := v.(string)
	return s, ok
}
