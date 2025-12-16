package hianime

type SeriesData struct {
	AnimeID      string `json:"anime_id"`
	EnglishName  string `json:"name"`
	AnilistID    string `json:"anilist_id"`
	SeriesUrl    string `json:"series_url"`
	JapaneseName string
}

type Episodes struct {
	Number        int
	EnglishTitle  string
	JapaneseTitle string
	Url           string
	Id            int
}

type ServerList struct {
	Type   string
	Name   string
	DataId int
	Id     int
}
type AjaxResponse struct {
	Status bool   `json:"status"`
	Html   string `json:"html"`
}

type MegacloudUrl struct {
	Type string `json:"type"`
	Url  string `json:"link"`
}

type StreamData struct {
	Url       string
	UserAgent string
	Referer   string
	Origin    string
	Tracks    []Track
	Intro     Timestamp
	Outro     Timestamp
}

type Track struct {
	File    string `json:"file"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Default bool   `json:"default"`
}

type Timestamp struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Sources struct {
	Sources   []Source  `json:"sources"`
	Tracks    []Track   `json:"tracks"`
	Encrypted bool      `json:"encrypted"`
	Intro     Timestamp `json:"intro"`
	Outro     Timestamp `json:"outro"`
	Server    int       `json:"server"`
}

type Source struct {
	File string `json:"file"`
	Type string `json:"type"`
}
