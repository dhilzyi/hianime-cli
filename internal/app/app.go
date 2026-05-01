package app

import (
	"bufio"
	"log"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/state"
)

type App struct {
	Config  *config.Config
	History []state.History
	Cache   *Cache
	AppDir  *config.AppPaths
	Flags   *Flags
	Scanner *bufio.Scanner
}

type Flags struct {
	Debug      bool
	Version    bool
	MpvVerbose bool
	Update     bool
}

func New(cfg *config.Config, history []state.History, appDir *config.AppPaths, flags *Flags, scanner bufio.Scanner) *App {
	return &App{
		Config:  cfg,
		History: history,
		AppDir:  appDir,
		Scanner: &scanner,
		Cache:   NewCache(),
		Flags:   flags,
	}
}

func (a *App) Start() {
	a.handleMenu()
}

func (a *App) SaveHistory(updated *state.History) {
	a.History = state.UpdateHistory(a.History, *updated)
	if err := state.SaveHistory(a.History, a.AppDir.DataDir); err != nil {
		log.Println("Failed to save history:", err)
	}
}
