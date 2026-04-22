package core

type Provider interface {
	Name() string
	GetEpisodes() ([]Episode, error)
	GetServers(string) ([]Server, error)
	GetStreamData(serverName string) (StreamData, error)
}
