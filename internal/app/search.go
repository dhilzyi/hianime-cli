package app

import (
	"fmt"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

type SearchState struct {
	provider core.Provider
	current  core.SearchPage
	cache    map[searchCacheKey]core.SearchPage
}

type searchCacheKey struct {
	query string
	page  int
}

func NewSearchState(p core.Provider) (*SearchState, error) {
	if p == nil {
		return nil, fmt.Errorf("provider is nil")
	}
	return &SearchState{
		provider: p,
		cache:    make(map[searchCacheKey]core.SearchPage),
	}, nil
}

func (s *SearchState) fetch(query string, page int) (core.SearchPage, error) {
	key := searchCacheKey{query, page}

	if cached, ok := s.cache[key]; ok {
		fmt.Println("Info: search cache hit")
		return cached, nil
	}

	result, err := s.provider.GetSearchResults(query, page)
	if err != nil {
		return core.SearchPage{}, err
	}

	s.cache[key] = result
	return result, nil
}

func (s *SearchState) Search(query string) error {
	page, err := s.fetch(query, 1)
	if err != nil {
		return err
	}
	s.current = page
	return nil
}

func (s *SearchState) Next() error {
	if !s.current.HasNext {
		return nil
	}
	page, err := s.fetch(s.current.Query, s.current.Page+1)
	if err != nil {
		return err
	}
	s.current = page
	return nil
}

func (s *SearchState) Prev() error {
	if !s.current.HasPrev {
		return nil
	}
	page, err := s.fetch(s.current.Query, s.current.Page-1)
	if err != nil {
		return err
	}
	s.current = page
	return nil
}

func (s *SearchState) Results() []core.SearchResult { return s.current.Results }
func (s *SearchState) HasNext() bool                { return s.current.HasNext }
func (s *SearchState) HasPrev() bool                { return s.current.HasPrev }
