package reanime

type seriesDataResponse struct {
	Anime    anime `json:"anime"`
	Episodes episodes
}

type anime struct {
	AnilistID     int    `json:"anilist_id"`
	AnimeID       string `json:"anime_id"`
	Duration      int    `json:"duration"`
	EpisodesTotal int    `json:"episodes_total"`
	StartDate     date   `json:"start_date"`
	Status        string
	Format        string
	Synonyms      []string
	Title         title
}

type episodes struct {
	Total      int
	TotalPages int `json:"totalPages"`

	Data []struct {
		Duration      int
		Title         string
		TitleRomaji   string `json:"title_romaji"`
		TitleJapanese string `json:"title_japanese"`
		Number        int    `json:"episode_number"`
	} `json:"data"`
}

type date struct {
	Day   int
	Month int
	Year  int
}

type title struct {
	English string
	Native  string
	Romaji  string
}

type serverApiResponse struct {
	Success bool
	Servers []struct {
		ID         string `json:"$id"`
		ServerName string `json:"serverName"`
		DataLink   string `json:"dataLink"`
		DataType   string `json:"dataType"`
		Continue   bool   `json:"continue"`
		Softsub    bool   `json:"softsub"`
	}
}

type searchApiResponse struct {
	Results []result
	Query   string
	Total   int
}

type result struct {
	AnimeID    string `json:"anime_id"`
	Title      title
	Status     string
	Format     string
	Episodes   int
	Duration   string
	SeasonYear int `json:"season_year"`
	Subbed     int
}

type query struct {
	Total      int
	LastOffset int
	rawQuery   string
}
