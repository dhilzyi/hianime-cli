package ui

import "fmt"

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
