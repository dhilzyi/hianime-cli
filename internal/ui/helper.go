package ui

import (
	"fmt"
	"strconv"
)

func prettyDuration(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func typeOrder(t string, listType []string) int {
	for i, typ := range listType {
		if typ == t {
			return i
		}
	}

	return len(listType)
}

func formatInt(val int) string {
	if val == 0 {
		return "N/A"
	}
	return strconv.Itoa(val)
}

func formatString(val string) string {
	if val == "" {
		return "N/A"
	}
	return val
}
