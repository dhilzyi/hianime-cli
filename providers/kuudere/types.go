package kuudere

type UrlType int

const (
	InvalidUrlType UrlType = iota
	AnimeType
	WatchType
)

// For front end response
type watchDataResponse struct {
	Success bool

	AnimeMetadata struct {
		ID           string      `json:"id"`
		AnilistID    int         `json:"anilist"`
		English      string      `json:"english"`
		Romaji       string      `json:"romaji"`
		Native       string      `json:"native"`
		AgeRating    string      `json:"ageRating"`
		MalScore     float64     `json:"malScore"`
		AverageScore int         `json:"averageScore"`
		Duration     int         `json:"duration"`
		Genres       []string    `json:"genres"`
		Cover        string      `json:"cover"`
		Banner       string      `json:"banner"`
		Season       string      `json:"season"`
		StartDate    string      `json:"startDate"`
		Status       string      `json:"status"`
		Synonyms     []string    `json:"synonyms"`
		Studios      []string    `json:"studios"`
		Type         string      `json:"type"`
		Year         int         `json:"year"`
		EpCount      int         `json:"epCount"`
		SubbedCount  int         `json:"subbedCount"`
		DubbedCount  int         `json:"dubbedCount"`
		Description  string      `json:"description"`
		Ep           int         `json:"ep"`
		UserLiked    bool        `json:"userLiked"`
		UserUnliked  bool        `json:"userUnliked"`
		Likes        int         `json:"likes"`
		Dislikes     int         `json:"dislikes"`
		InWatchlist  bool        `json:"inWatchlist"`
		Folder       interface{} `json:"folder"`
		URL          string      `json:"url"`
	} `json:"anime_info"`
}

type animeDataResponse struct {
	Success bool `json:"success"`

	AnimeMetadata struct {
		ID             string   `json:"id"`
		AnilistID      int      `json:"anilistId"`
		English        string   `json:"english"`
		Romaji         string   `json:"romaji"`
		Native         string   `json:"native"`
		AgeRating      string   `json:"ageRating"`
		MalScore       float64  `json:"malScore"`
		AverageScore   int      `json:"averageScore"`
		Duration       int      `json:"duration"`
		Studios        []string `json:"studios"`
		Genres         []string `json:"genres"`
		Cover          string   `json:"cover"`
		Banner         string   `json:"banner"`
		Season         string   `json:"season"`
		StartDate      string   `json:"startDate"`
		Status         string   `json:"status"`
		Synonyms       []string `json:"synonyms"`
		Type           string   `json:"type"`
		Year           int      `json:"year"`
		EpCount        int      `json:"epCount"`
		SubbedCount    int      `json:"subbedCount"`
		DubbedCount    int      `json:"dubbedCount"`
		Description    string   `json:"description"`
		Views          string   `json:"views"`
		Likes          string   `json:"likes"`
		InWatchlist    bool     `json:"in_watchlist"`
		RelatedSeasons []struct {
			ID           string `json:"id"`
			Title        string `json:"title"`
			Cover        string `json:"cover"`
			Type         string `json:"type"`
			Year         int    `json:"year"`
			Status       string `json:"status"`
			EpisodeCount int    `json:"episodeCount"`
		} `json:"relatedSeasons"`
		ContinueWatching interface{} `json:"continueWatching"`
		CurrentAnimeID   string      `json:"currentAnimeId"`
	} `json:"data"`
}

type episodesResponse struct {
	Status string `json:"status"`
	Data   []struct {
		AccessID  string      `json:"access_id"`
		Audio     string      `json:"audio"`
		AnilistID int         `json:"anilist_id"`
		MalID     interface{} `json:"mal_id"`
		Episode   int         `json:"episode"`
		PlayerURL string      `json:"player_url"`
	} `json:"data"`
	Pagination struct {
		CurrentPage int         `json:"current_page"`
		PerPage     int         `json:"per_page"`
		TotalItems  int         `json:"total_items"`
		TotalPages  int         `json:"total_pages"`
		NextPage    interface{} `json:"next_page"`
		PrevPage    interface{} `json:"prev_page"`
	} `json:"pagination"`
}

type searchResponse struct {
	Success   bool `json:"success"`
	Total     int  `json:"total"`
	Documents []struct {
		ID           string   `json:"id"`
		English      string   `json:"english"`
		Romaji       string   `json:"romaji"`
		Native       string   `json:"native"`
		AgeRating    string   `json:"ageRating"`
		MalScore     float64  `json:"malScore"`
		AverageScore int      `json:"averageScore"`
		Duration     int      `json:"duration"`
		Genres       []string `json:"genres"`
		Cover        string   `json:"cover"`
		Season       string   `json:"season"`
		StartDate    string   `json:"startDate"`
		Status       string   `json:"status"`
		Synonyms     []string `json:"synonyms"`
		Type         string   `json:"type"`
		Year         int      `json:"year"`
		EpCount      int      `json:"epCount"`
		SubbedCount  int      `json:"subbedCount"`
		DubbedCount  int      `json:"dubbedCount"`
		Description  string   `json:"description"`
	} `json:"documents"`
	TotalPages int `json:"totalPages"`
}
