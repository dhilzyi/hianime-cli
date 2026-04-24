package kuudere

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/core"
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

type TokenApiResponse struct {
	VideoB64 string `json:"video_b64"`
	KeyFrag  string `json:"key_frag"`
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

func buildCryptoData(data RawJson5) (*CryptoData, error) {
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

	key, ok := getString(object[dynamic["key_field"]])
	if !ok {
		return nil, fmt.Errorf("key not in map")
	}

	iv, ok := getString(object[dynamic["iv_field"]])
	if !ok {
		return nil, fmt.Errorf("iv not in map")
	}

	secondaryKey, ok := getString(data.RawData[dynamic["secondary_key_field"]])
	if !ok {
		return nil, fmt.Errorf("secondaryKey is not in the data")
	}
	token, ok := getString(data.RawData[dynamic["token_field"]])
	if !ok {
		return nil, fmt.Errorf("token is not in the data")
	}

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

func fetchTokenApi(tokenReference, defaultDomain string, client *http.Client) (TokenApiResponse, error) {
	url := fmt.Sprintf("%sapi/m3u8/%s/", defaultDomain, tokenReference)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TokenApiResponse{}, err
	}
	fmt.Println(url)
	resp, err := client.Do(req)
	if err != nil {
		return TokenApiResponse{}, err
	}
	defer resp.Body.Close()

	var tokenResp TokenApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return TokenApiResponse{}, err
	}

	return tokenResp, nil
}

func ResolveZerocloudz(rawUrl string) (core.StreamData, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return core.StreamData{}, err
	}
	defaultDomain := fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host)
	client, err := NewSession(defaultDomain)
	if err != nil {
		return core.StreamData{}, err
	}

	rawHtml, err := fetchRawHtml(rawUrl, client)
	if err != nil {
		return core.StreamData{}, err
	}

	extracted, err := ExtractPayload(rawHtml)
	if err != nil {
		return core.StreamData{}, err
	}
	json5Data, err := ParseToJson5(extracted)
	if err != nil {
		return core.StreamData{}, err
	}

	cryptoData, err := buildCryptoData(json5Data)
	if err != nil {
		return core.StreamData{}, err
	}

	tokenApiResp, err := fetchTokenApi(cryptoData.Token, defaultDomain, client)
	if err != nil {
		return core.StreamData{}, err
	}

	encryptedKeyBytes, err := cleanBase64(cryptoData.Key)
	if err != nil {
		return core.StreamData{}, err
	}

	secondaryKeyBytes, err := cleanBase64(cryptoData.Extra)
	if err != nil {
		return core.StreamData{}, err
	}

	dynamicKeyBytes, err := cleanBase64(tokenApiResp.KeyFrag)
	if err != nil {
		return core.StreamData{}, err
	}

	ivBytes, err := cleanBase64(cryptoData.IV)
	if err != nil {
		return core.StreamData{}, err
	}
	ciphertextBytes, err := cleanBase64(tokenApiResp.VideoB64)
	if err != nil {
		return core.StreamData{}, err
	}

	videoUrl, err := PerformDecryption(json5Data.Data.ObfuscationSeed, encryptedKeyBytes, secondaryKeyBytes, dynamicKeyBytes, ivBytes, ciphertextBytes)
	if err != nil {
		return core.StreamData{}, err
	}

	return core.StreamData{
		Url: videoUrl,
	}, nil
}

func fetchRawHtml(inputUrl string, client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", inputUrl, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	htmlRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(htmlRaw), nil
}

// GenerateSBox creates a simple substitution box (S-box) for key derivation
func GenerateSBox(seed int) []byte {
	sbox := make([]byte, 256)

	for i := 0; i < 256; i++ {
		sbox[i] = byte((i*37 + seed) & 0xFF)
	}

	return sbox
}

func DeriveAESKey(keyFragment, secondaryKey, dynamicKey, sbox []byte) ([]byte, error) {
	length := len(keyFragment)

	// Crash-proof safety check!
	if len(secondaryKey) < length || len(dynamicKey) < length {
		return nil, fmt.Errorf("secondaryKey or dynamicKey is too short")
	}

	aesKey := make([]byte, length)

	for i := 0; i < length; i++ {
		aesKey[i] = keyFragment[i] ^ secondaryKey[i] ^ dynamicKey[i] ^ sbox[i&0xFF]
	}

	return aesKey, nil
}

func PerformDecryption(
	obfuscationSeed string,
	encryptedKeyBytes, secondaryKeyBytes, dynamicKeyBytes, ivBytes, ciphertextBytes []byte,
) (string, error) {

	// 1. Generate S-box seed (int(obfuscation_seed[0:8], 16))
	seedHex := obfuscationSeed[0:8]
	sboxSeedInt64, err := strconv.ParseInt(seedHex, 16, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse seed hex: %w", err)
	}
	sboxSeed := int(sboxSeedInt64)

	// 2. Generate S-box and AES Key
	sboxTable := GenerateSBox(sboxSeed)

	aesKey, err := DeriveAESKey(encryptedKeyBytes, secondaryKeyBytes, dynamicKeyBytes, sboxTable)
	if err != nil {
		return "", fmt.Errorf("failed to derive AES key: %w", err)
	}

	// 3. Decrypt AES-256-CBC
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	if len(ciphertextBytes)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// Create decrypter and decrypt into a new byte slice
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	plaintextBytes := make([]byte, len(ciphertextBytes))
	mode.CryptBlocks(plaintextBytes, ciphertextBytes)

	// 4. Unpad (equivalent to unpad(plaintext_bytes, AES.block_size))
	unpaddedBytes, err := unpadPKCS7(plaintextBytes)
	if err != nil {
		return "", fmt.Errorf("failed to unpad decrypted data: %w", err)
	}

	// .decode() -> In Go, converting []byte to string handles UTF-8 automatically
	return string(unpaddedBytes), nil
}

// unpadPKCS7 removes standard AES padding
func unpadPKCS7(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("empty data")
	}

	// In PKCS7, the last byte tells us how many bytes of padding were added
	paddingLen := int(data[length-1])

	// Sanity check to ensure padding length is valid
	if paddingLen > length || paddingLen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	// Slice off the padding
	return data[:length-paddingLen], nil
}

func cleanBase64(input string) ([]byte, error) {
	// 1. Remove hidden newlines, spaces, and carriage returns
	clean := strings.TrimSpace(input)

	// 2. Remove accidental leftover JSON quotes just in case
	clean = strings.Trim(clean, `"`)

	// 3. Decode
	bytes, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to decode clean b64 '%s': %w", clean, err)
	}

	return bytes, nil
}
