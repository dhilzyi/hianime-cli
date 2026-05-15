package anikoto

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/dhilzyi/hianime-cli/hosts/megaplay"
	"github.com/dhilzyi/hianime-cli/internal/common"
	"github.com/dhilzyi/hianime-cli/internal/core"
)

const (
	baseURL   = "https://anikototv.to"
	userAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:150.0) Gecko/20100101 Firefox/150.0`
)

type Anikoto struct {
	episodes map[int]episode
	servers  map[string]server
	inputURL string
	session  *session
}

func New(rawURL string) (*Anikoto, error) {
	client, err := common.NewSession()
	if err != nil {
		return nil, err
	}
	return &Anikoto{
		inputURL: rawURL,
		episodes: make(map[int]episode),
		servers:  make(map[string]server),
		session: &session{
			http: client,
		},
	}, nil
}

func (a *Anikoto) Name() string {
	return "Anikoto"
}

func (a *Anikoto) GetEpisodes() ([]core.Episode, *core.SeriesData, error) {
	seriesData, err := a.session.getSeriesData(a.inputURL)
	if err != nil {
		return nil, nil, err
	}
	episodesData, err := a.session.getEpisodes(seriesData.animeID)
	if err != nil {
		return nil, nil, err
	}
	a.episodes = episodesData
	var epsCore []core.Episode
	for _, e := range episodesData {
		epsCore = append(epsCore, e.Episode)
	}
	sort.Slice(epsCore, func(i, j int) bool {
		return epsCore[i].Number < epsCore[j].Number
	})

	return epsCore, &seriesData.SeriesData, nil
}

func (a *Anikoto) GetServers(episode core.Episode) ([]core.Server, error) {
	serversData, err := a.session.getServers(a.episodes[episode.Number].serverDataId)
	if err != nil {
		return nil, err
	}
	a.servers = serversData
	var srvCore []core.Server
	for _, srv := range serversData {
		srvCore = append(srvCore, srv.Server)
	}
	sort.SliceStable(srvCore, func(i, j int) bool {
		typeRank := map[string]int{"SUB": 0, "DUB": 1}
		return typeRank[srvCore[i].Type] < typeRank[srvCore[j].Type]
	})
	return srvCore, nil
}

func (a *Anikoto) GetStreamData(serverKey string) (core.StreamData, error) {
	srv, exists := a.servers[serverKey]
	if !exists || srv.DataLinkId == "" {
		return core.StreamData{}, fmt.Errorf("could not found server entry: %s", serverKey)
	}
	serverURL, err := a.session.getServerUrl(srv.DataLinkId)
	if err != nil {
		return core.StreamData{}, err
	}
	var streamData core.StreamData
	if strings.Contains(serverURL, "megaplay") || strings.Contains(serverURL, "vidwish") {
		streamData, err = megaplay.GetStreamData(serverURL, baseURL)
		if err != nil {
			return core.StreamData{}, err
		}
	} else {
		return core.StreamData{}, fmt.Errorf("server url is not supported: %s", serverURL)
	}

	return streamData, nil
}

func (a *Anikoto) GetSearchResults(rawQuery string, page int) (core.SearchPage, error) {
	query := common.StringToQueryFormat(rawQuery)
	searchURL := fmt.Sprintf("%s/filter?keyword=%s&page=%d", baseURL, query, page)
	searchResult, err := a.session.getSearch(searchURL)
	if err != nil {
		return core.SearchPage{}, err
	}
	searchResult.Query = rawQuery

	return *searchResult, nil
}

func (a *Anikoto) ExtractProviderID() (string, error) {
	u, err := url.Parse(a.inputURL)
	if err != nil {
		return "", err
	}
	path := strings.Trim(u.Path, "/")
	path = strings.TrimPrefix(path, "watch")
	splitted := strings.Split(path, "/")
	var id string
	if len(splitted) > 1 {
		id = splitted[1]
	} else {
		id = splitted[0]
	}
	return id, nil
}
