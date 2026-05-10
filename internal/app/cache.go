package app

import (
	"github.com/dhilzyi/hianime-cli/internal/core"
)

type CacheEntry struct {
	SeriesData core.SeriesData
	Episodes   []core.Episode
}

type Cache struct {
	byProviderID map[string]*CacheEntry
	byAnilistID  map[int]*CacheEntry
}

func NewCache() *Cache {
	return &Cache{
		byProviderID: make(map[string]*CacheEntry),
		byAnilistID:  make(map[int]*CacheEntry),
	}
}
