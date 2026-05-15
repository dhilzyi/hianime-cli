package core

// Series data anime
type SeriesData struct {
	Titles Title

	// from 'anilist.co' for jimaku and other things.
	AnilistID int    `json:"anilist_id"`
	SeriesUrl string `json:"series_url"`
}

// Use for building commands mpv to play.
type StreamData struct {
	Url     string
	Headers map[string]string

	Chapters []Timestamp
	Tracks   []Track
}

// Intro, Outro and other chapters
type Timestamp struct {
	Start int
	End   int
	Name  string
}

// Subtitles or thumbnails
type Track struct {
	Url  string
	Name string
	Type string

	// Optional
	Language string
}

// Individual episode struct for list of series
type Episode struct {
	Url string

	Number int
	Titles Title
}

// Title support three types
type Title struct {
	RomajiTitle  string
	EnglishTitle string
	KanjiTitle   string
}

// Contains server list
type Server struct {
	Name string
	Type string
	Key  string
}

type SearchResult struct {
	Titles         Title
	Type           string
	NumberEpisodes int
	Duration       int
	Url            string
	Year           int
}

type SearchPage struct {
	// Results for data and ui print
	Results []SearchResult

	// Current page number for next and prev if available
	Page    int
	HasNext bool
	HasPrev bool

	// Raw input query for key
	Query string
}
