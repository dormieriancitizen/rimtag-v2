package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

func GetTSV(mods []*Mod) string {
	mods_csv := [][]string{}
	for _, mod := range mods {
		mods_csv = append(mods_csv, mod.TSVInfo())
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = '\t' // use tab as the separator

	err := writer.WriteAll(mods_csv) // writes all rows at once
	if err != nil {
		fmt.Println("Error writing CSV:", err)
		return ""
	}

	return buf.String()
}

func ExportMarkdown(mods []*Mod, config Config) string {
	md := fmt.Sprintf(
		`# RimWorld Mod List: %[1]d mods       ![](https://github.com/RimSort/RimSort/blob/main/docs/rentry_preview.png?raw=true)
Created with a bad go script with a lot of borrowed code from RimSort
Mod list was created for game version: %s
!!! note Mod list length:`+"`%[1]d`"+`
***
# | Mod Name | Info
:-: | ------ | :------:`,
		len(mods), GetRimworldVersion(config),
	)
	modLines := []string{}
	for i, mod := range mods {
		line := "\n"
		line += fmt.Sprint(i)
		line += "|"

		var banner string
		switch mod.Source {
		case ModSourceLocal:
			banner = "https://github.com/dormieriancitizen/rimworld_instance_manager/blob/main/resources/local-banner.png?raw=true"
		case ModSourceGit:
			banner = "https://github.com/dormieriancitizen/rimworld_instance_manager/blob/main/resources/github-banner.png?raw=true"
		case ModSourceOfficial:
			banner = "https://github.com/dormieriancitizen/rimworld_instance_manager/blob/main/resources/ludeon-studios.png?raw=true"
		case ModSourceSteam:
			if mod.SteamInfo != nil {
				if len(mod.SteamInfo.PreviewURL) > 0 {
					banner = mod.SteamInfo.PreviewURL + "?imw=100&imh=100&impolicy=Letterbox"
				} else {
					banner = "https://github.com/RimSort/RimSort/blob/main/docs/rentry_steam_icon.png?raw=true"
				}
			} else {
				banner = "https://github.com/RimSort/RimSort/blob/main/docs/rentry_steam_icon.png?raw=true"
			}
		}

		line += fmt.Sprintf("![%s](%s){100px:56px} ", mod.PackageID, banner)

		var name string
		if len(mod.About.Name) > 0 {
			name = mod.About.Name
		} else {
			name = string(mod.PackageID)
		}

		if mod.Source == ModSourceSteam {
			url := "https://steamcommunity.com/sharedfiles/filedetails/?id=" + fmt.Sprint(mod.GetPublishedAppID())
			line += fmt.Sprintf("[%s](%s)", name, url)
		} else if len(mod.About.Url) != 0 {
			line += fmt.Sprintf("[%s](%s)", name, mod.About.Url)
		} else {
			line += name
		}
		if mod.Source == ModSourceLocal {
			line += " {packageid: " + string(mod.PackageID) + "}"
		}

		line += " | " + string(mod.BestSupportedVersion())

		modLines = append(modLines, line)
	}
	md += strings.Join(modLines, "")

	// "\n***"
	// "\n# | Mod Name | Info"
	// "\n:-: | ------ | :------:"
	return md
}
