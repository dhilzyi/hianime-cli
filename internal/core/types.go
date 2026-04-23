package core

// Series data anime
type SeriesData struct {
	Titles    Title
	AnilistID int    `json:"anilist_id"`
	SeriesUrl string `json:"series_url"`
}

// Use for building commands for mpv to play.
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
}

// Subtitles or thumbnails
type Track struct {
	// Url of the track
	File string

	// Name of the track
	Label string

	// Type track
	Kind string
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
}
