package main

import (
	"encoding/xml"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ModConfig struct {
	XMLName         xml.Name `xml:"ModsConfigData"`
	Version         string   `xml:"version"`
	KnownExpansions []string `xml:"knownExpansions>li"`
	ActiveMods      []string `xml:"activeMods>li"`
}

func SetModlist(mods []*Mod, config Config) error {
	rimworldVersion := GetRimworldVersion(config)
	expansions, err := GetRimworldExpansions(config)
	if err != nil {
		return err
	}

	knownExpansions := []string{}
	for _, expansion := range expansions {
		if expansion.PackageID == "Ludeon.RimWorld" {
			continue
		}
		knownExpansions = append(knownExpansions, string(expansion.PackageID))
	}

	activeMods := []string{}
	for _, mod := range mods {
		activeMods = append(activeMods, string(mod.PackageID))
	}

	modConfig := ModConfig{
		Version:         rimworldVersion,
		KnownExpansions: knownExpansions,
		ActiveMods:      activeMods,
	}
	modsConfigXml, err := xml.MarshalIndent(modConfig, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(config.RimworldData, "Config/ModsConfig.xml"), modsConfigXml, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadModlist(mods []*Mod, config Config) error {
	LinkMods(mods)
	err := CheckDeps(mods, config)
	if err != nil {
		return err
	}

	err = SymlinkMods(mods, config)
	if err != nil {
		return err
	}
	sortedMods, err := SortMods(mods)
	if err != nil {
		return err
	}

	SetModlist(sortedMods, config)
	return nil
}

func GetModsFromPath(path string, config Config) ([]*Mod, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	mods := make([]*Mod, 0, len(lines))
	for i, path := range lines {
		if path == "" {
			continue
		}

		mod, err := ParseMod(path, config)
		if err != nil {
			log.Printf("ParseMod failed at line %d (%s): %v", i+1, path, err)
			continue
		}

		mods = append(mods, mod)
	}
	return mods, nil
}
