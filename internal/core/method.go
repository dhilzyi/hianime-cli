package core

func (t Title) GetPreferredTitle() string {
	var finalTitle string
	if t.KanjiTitle != "" {
		finalTitle = t.KanjiTitle
	}
	if t.EnglishTitle != "" {
		finalTitle = t.EnglishTitle
	}
	if t.RomajiTitle != "" {
		finalTitle = t.RomajiTitle
	}
	if finalTitle == "" {
		finalTitle = "UNKNOWN"
	}

	return finalTitle
}
