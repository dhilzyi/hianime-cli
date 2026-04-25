package kuudere

// For front end response
type seriesDataResponse struct {
	AnimeInfo struct {
		ID           string      `json:"id"`
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
		Anilist      int         `json:"anilist"`
		URL          string      `json:"url"`
	} `json:"anime_info"`
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
