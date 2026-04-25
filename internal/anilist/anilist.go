package anilist

import (
	"fmt"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

func FillSeriesData(seriesData *core.SeriesData) error {
	var graphresp GraphQLResponse
	var err error

	if seriesData.AnilistID != 0 {
		graphresp, err = getAnilistDataById(seriesData.AnilistID)
		if err != nil {
			return fmt.Errorf("failed to fetch from Anilist by ID: %w", err)
		}
	} else {
		var title string
		if seriesData.Titles.EnglishTitle != "" {
			title = seriesData.Titles.EnglishTitle
		} else if seriesData.Titles.RomajiTitle != "" {
			title = seriesData.Titles.RomajiTitle
		} else if seriesData.Titles.KanjiTitle != "" {
			title = seriesData.Titles.KanjiTitle
		} else {
			return fmt.Errorf("error: All titles and AnilistID are empty, cannot query")
		}

		graphresp, err = getAnilistData(title)
		if err != nil {
			return fmt.Errorf("failed to fetch from Anilist by Title: %w", err)
		}
	}

	if seriesData.AnilistID == 0 {
		seriesData.AnilistID = graphresp.Data.Media.ID
	}
	if seriesData.Titles.EnglishTitle == "" {
		seriesData.Titles.EnglishTitle = graphresp.Data.Media.Title.English
	}
	if seriesData.Titles.RomajiTitle == "" {
		seriesData.Titles.RomajiTitle = graphresp.Data.Media.Title.Romaji
	}
	if seriesData.Titles.KanjiTitle == "" {
		seriesData.Titles.KanjiTitle = graphresp.Data.Media.Title.Native
	}

	return nil
}
