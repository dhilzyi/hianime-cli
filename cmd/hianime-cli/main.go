package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/release"
	"github.com/dhilzyi/hianime-cli/internal/state"
)

type InputType int

const (
	InputURL InputType = iota
	InputHistoryIndex
	InputCommand
	InputAnilistID
)

//go:embed version.txt
var embedVersion string

func main() {
	cleanEmbedVersion := strings.TrimSpace(embedVersion)
	flags := ParseFlags()
	HandleFlags(flags, cleanEmbedVersion)

	scanner := bufio.NewScanner(os.Stdin)

	appDir, err := config.InitPath()
	if err != nil {
		fmt.Println("Fail to initialize path: " + err.Error())
	}
	configSes, err := config.LoadConfig(appDir.ConfigDir, cleanEmbedVersion)
	if err != nil {
		fmt.Println("Fail to load config file: " + err.Error())
	}

	if newCfg, updated, err := release.Run(cleanEmbedVersion, appDir.DataDir, configSes); err != nil {
		fmt.Println("Fail to run version module:", err)
	} else {
		if updated {
			configSes = newCfg
			fmt.Printf("Set to new config complete.\n")
		}
	}

	history, err := state.LoadHistory(appDir.DataDir)
	if err != nil {
		fmt.Println("Fail to load history file: " + err.Error())
	}

	// Using anilistID as a key for the cache
	// and using unique clean path from url itself
	cache := Cache{
		byProviderID: make(map[string]*CacheEntry),
		byAnilistID:  make(map[int]*CacheEntry),
	}

	app := &App{
		Config:  &configSes,
		History: history,
		Cache:   &cache,
		AppDir:  &appDir,
		Flags:   &flags,
		Scanner: scanner,
	}

	app.handleMenu(history)
}
