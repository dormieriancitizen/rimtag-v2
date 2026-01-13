package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	SteamModSrc  string `toml:"steam-src"    comment:"Source directory where steam mods can be found (often Steam/steamapps/workshop/content/294100)"`
	LocalModSrc  string `toml:"local-src"    comment:"Source directory where local and git mods can be found"`
	RimworldData string `toml:"rimworld-data" comment:"Rimworld's data path (often .config/unity3d/Ludeon Studios/RimWorld by Ludeon Studios/)"`
	TargetDir    string `toml:"target-dir" comment:"Rimworld path"`
}

func getConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configHome = filepath.Join(homeDir, ".config")
		} else {
			configHome = ".config"
		}
	}
	return filepath.Join(configHome, "rimtag")
}

func LoadConfig() Config {
	configRoot := getConfigPath()
	configPath := filepath.Join(configRoot, "config.toml")
	data, err := os.ReadFile(configPath)

	var cfg Config
	if errors.Is(err, os.ErrNotExist) {
		os.Mkdir(configRoot, 0644)
		cfg = Config{
			SteamModSrc:  "/home/dormierian/.local/share/Steam/steamapps/workshop/content/294100",
			LocalModSrc:  "/home/dormierian/.config/rimtag/mods",
			RimworldData: "/home/dormierian/.config/unity3d/Ludeon Studios/RimWorld by Ludeon Studios/",
			TargetDir:    "/home/dormierian/Games/rimworld",
		}
		data, err := toml.Marshal(cfg)
		if err != nil {
			fmt.Println("Failed to marshal default config")
			panic(err)
		}
		err = os.WriteFile(configPath, data, 0644)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	_, err = os.Lstat(filepath.Join(configRoot, "mods"))
	if err != nil {
		os.Mkdir(filepath.Join(configRoot, "mods"), 0644)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
