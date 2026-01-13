package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetRimworldVersion(config Config) string {
	content, err := os.ReadFile(filepath.Join(config.TargetDir, "Version.txt"))
	if err != nil {
		return "1.6.4633 rev1273"
	}
	return strings.TrimRight(string(content), "\r\n")
}

func GetRimworldMajorVersion(config Config) string {
	return GetRimworldVersion(config)[:3]
}

func GetRimworldExpansions(config Config) ([]*Mod, error) {
	expansionPath := filepath.Join(config.TargetDir, "Data")
	modPaths, err := os.ReadDir(expansionPath)
	if err != nil {
		return nil, err
	}

	expansions := []*Mod{}
	for _, entry := range modPaths {
		if !entry.IsDir() {
			continue
		}
		subdirPath := filepath.Join(expansionPath, entry.Name())
		expansion, err := ParseMod(subdirPath, config)
		if err != nil {
			fmt.Printf("Unsuccessful parse on %s @ %s\n", entry.Name(), subdirPath)
		}
		expansions = append(expansions, expansion)
	}
	return expansions, nil
}

func SymlinkMods(mods []*Mod, config Config) error {
	modsDir := filepath.Join(config.TargetDir, "Mods")

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(modsDir, entry.Name())
		info, err := os.Lstat(path)
		if err != nil {
			fmt.Println("Error getting info for", path, ":", err)
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(path); err != nil {
				fmt.Println("Failed to unlink mod: ", err)
			}
		}
	}

	for _, mod := range mods {
		if mod.Source == ModSourceOfficial {
			continue
		}
		err := os.Symlink(mod.Path, filepath.Join(config.TargetDir, "Mods/"+string(mod.PackageID)))
		if err != nil {
			log.Printf("Failed to link %s\n", mod)
			log.Println(err)
		}
	}

	return nil
}
