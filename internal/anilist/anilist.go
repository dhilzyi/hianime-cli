package anilist

import (
	"fmt"

	"github.com/dhilzyi/hianime-cli/internal/core"
)

func FillSeriesData(seriesData *core.SeriesData) error {
	var title string
	if seriesData.Titles.EnglishTitle != "" {
		title = seriesData.Titles.EnglishTitle
	} else if seriesData.Titles.RomajiTitle != "" {
		title = seriesData.Titles.RomajiTitle
	} else if seriesData.Titles.KanjiTitle != "" {
		title = seriesData.Titles.KanjiTitle
	} else {
		return fmt.Errorf("error: All titles are empty and query can not be process")
	}

	graphresp, err := getAnilistData(title)
	if err != nil {
		return err
	}
	seriesData.AnilistID = graphresp.Data.Media.ID
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
