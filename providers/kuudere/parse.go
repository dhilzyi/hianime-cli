package kuudere

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

// GenerateSBox creates a simple substitution box (S-box) for key derivation
func generateSBox(seed int) []byte {
	sbox := make([]byte, 256)

	for i := 0; i < 256; i++ {
		sbox[i] = byte((i*37 + seed) & 0xFF)
	}

	return sbox
}

func deriveAESKey(keyFragment, secondaryKey, dynamicKey, sbox []byte) ([]byte, error) {
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

func performDecryption(
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
	sboxTable := generateSBox(sboxSeed)

	aesKey, err := deriveAESKey(encryptedKeyBytes, secondaryKeyBytes, dynamicKeyBytes, sboxTable)
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
	clean := strings.TrimSpace(input)
	clean = strings.Trim(clean, `"`)

	bytes, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to decode clean b64 '%s': %w", clean, err)
	}

	return bytes, nil
}

func extractPayload(html string) (string, error) {
	pattern := `data:\s*\[null,null,(\{.*?\})\],\s*form:\snull`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 1 {
		return "", fmt.Errorf("could not find SvelteKit data payload")
	}

	jsonPayload := matches[1]

	return jsonPayload, nil
}

func parseToJson5(raw string) (rawJson5, error) {
	var result rawJson5
	if err := json5.Unmarshal([]byte(raw), &result); err != nil {
		return rawJson5{}, err
	}

	var fullMap map[string]interface{}
	json5.Unmarshal([]byte(raw), &fullMap)
	if dataMap, ok := fullMap["data"].(map[string]interface{}); ok {
		result.RawData = dataMap
	}

	return result, nil
}

func buildCryptoData(data rawJson5) (*cryptoData, error) {
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

	result := &cryptoData{
		Key:   key,
		IV:    iv,
		Token: token,
		Extra: secondaryKey,
	}

	return result, nil
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
