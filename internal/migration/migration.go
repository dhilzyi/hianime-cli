package migration

import (
	"fmt"
	"log"

	"github.com/dhilzyi/hianime-cli/internal/config"
	"github.com/dhilzyi/hianime-cli/internal/player"
)

func Run(dataDir, embedVer string, oldCfg config.Settings) (config.Settings, error) {
	fmt.Println("Update and migration started...")
	scriptPath := player.TrackScriptPath(dataDir, player.ScriptName)
	if err := player.WriteLuaScript(scriptPath); err != nil {
		log.Println(err)
	} else {
		fmt.Println("	Update lua script completed.")
	}
	newCfg, err := config.MigrateConfig(oldCfg, embedVer)
	if err != nil {
		return config.Settings{}, err
	}
	fmt.Println("	Migrate to new config completed.")

	return newCfg, nil
}
