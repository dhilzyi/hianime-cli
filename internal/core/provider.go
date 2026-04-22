package core

type Provider interface {
	Name() string
	GetEpisodes() ([]Episode, *SeriesData, error)
	GetServers(Episode) ([]Server, error)
	GetStreamData(serverName string) (StreamData, error)
}
