package anikoto

import (
	"strconv"
	"strings"
)

type ajaxResponse struct {
	Status int
	Result string
}
type ajaxServerResponse struct {
	Status int
	Result struct {
		Url      string
		SkipData struct {
			Intro []int
			Outro []int
		} `json:"skip_data"`
	}
}

func toValidHtml(rawHtml string) (string, error) {
	clean, err := strconv.Unquote(`"` + rawHtml + `"`)
	if err != nil {
		// fallback: manual replace
		clean = strings.NewReplacer(
			`\n`, "\n",
			`\t`, "\t",
			`\"`, `"`,
			`\\`, `\`,
		).Replace(rawHtml)
	}

	return clean, nil
}
