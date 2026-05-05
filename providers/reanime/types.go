package reanime

type seriesData struct {
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
