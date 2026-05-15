package core

// Each of Provider struct must be satisfy this interface
type Provider interface {
	// The name of provider
	Name() string

	// Retrieve Episode data and Series data
	GetEpisodes() ([]Episode, *SeriesData, error)

	// Retrieve Servers data
	GetServers(episode Episode) ([]Server, error)

	// Retrieve Stream data for mpv
	GetStreamData(serverName string) (StreamData, error)

	// Retrieve search results
	GetSearchResults(query string, page int) (SearchPage, error)

	// This id is for key cache purpose
	ExtractProviderID() (string, error)
}
