package core

type Provider interface {
	Name() string
	CanHandle(url string) bool
	GetEpisodeList(url string) ([]EpisodeList, error)
}
