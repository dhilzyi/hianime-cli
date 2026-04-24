package kuudere

import (
	"crypto/sha256"
	"encoding/hex"
)

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
