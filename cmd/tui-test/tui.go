package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/app"
	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/core"
	"github.com/dhilzyi/hianime-cli/internal/state"

	tea "charm.land/bubbletea/v2"
)

type screenState int

const (
	StateHistory screenState = iota
	StateEpisodes
	StatePlaying

	StateConfigSettings
)

type model struct {
	state  screenState
	cursor int
	subs   bool

	session *session
	cfg     *config.Config
}

func (m model) Init() tea.Cmd {
	return nil
}

type session struct {
	urlSeries string

	provider   core.Provider
	seriesData core.SeriesData

	historyList []state.History
	episodeList []core.Episode
	serverList  []core.Episode
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.state {
		case StateHistory:
			switch msg.String() {
			case "q", "esc":
				return m, tea.Quit
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.session.historyList) - 1
				}
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.session.historyList) {
					m.cursor = 0
				}
			case "s":
				m.subs = !m.subs
			case "c":
				m.state = StateConfigSettings
			case "enter":
				m.session.urlSeries = m.session.historyList[m.cursor].Metadata.SeriesUrl
				fetchResult, err := app.ResolveInput(app.ResolveParams{URL: m.session.urlSeries}, app.NewCache())
				if err != nil {
					log.Fatal(err)
					return m, tea.Quit
				}
				m.session.provider = fetchResult.Provider
				m.session.episodeList = fetchResult.Episodes
				m.session.seriesData = fetchResult.SeriesData

				m.state = StateEpisodes
				m.cursor = 0
			}
		case StateEpisodes:
			switch msg.String() {
			case "q", "esc":
				m.state = StateHistory
				m.cursor = 0
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.session.episodeList) - 1
				}
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.session.episodeList) {
					m.cursor = 0
				}
			case "enter":
			}
		}

	}

	return m, nil
}

func (m model) View() tea.View {
	var buf strings.Builder

	switch m.state {
	case StateHistory:
		buf.WriteString("--- RECENT HISTORY ---\n\n")
		for i, item := range m.session.historyList {
			if m.cursor == i {
				buf.WriteString("-> ")
			} else {
				buf.WriteString("   ")
			}
			buf.WriteString(item.Metadata.Titles.GetPreferredTitle() + "\n")
		}
		fmt.Fprintf(&buf, "\n\n Subs: %t\n", m.subs)
	case StateEpisodes:
		buf.WriteString("--- EPISODES ---\n\n")
		for i, ep := range m.session.episodeList {
			if m.cursor == i {
				buf.WriteString("-> ")
			} else {
				buf.WriteString("   ")
			}
			fmt.Fprintf(&buf, fmt.Sprintf("[%d] %s\n", ep.Number, ep.Titles.GetPreferredTitle()))
		}
		fmt.Fprintf(&buf, "\nSelected: %s", m.session.episodeList[m.cursor].Titles.GetPreferredTitle())
	}

	return tea.NewView(buf.String())
}
