package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v3"
)

func CmdVanilla(context.Context, *cli.Command) error {
	config := LoadConfig()
	mods, err := GetRimworldExpansions(config)
	if err != nil {
		return err
	}
	return LoadModlist(mods, config)
}
func CmdTsv(context.Context, *cli.Command) error {
	config := LoadConfig()
	fmt.Println(GetTSV(config))
	return nil
}
func CmdLoad(context.Context, *cli.Command) error {
	config := LoadConfig()
	filename := "listpaths.tsv"

	mods, err := GetModsFromPath(filename, config)
	if err != nil {
		return err
	}

	return LoadModlist(mods, config)
}
func CmdMarkdown(context.Context, *cli.Command) error {
	config := LoadConfig()
	filename := "listpaths.tsv"

	mods, err := GetModsFromPath(filename, config)
	if err != nil {
		return err
	}
	LinkMods(mods)
	AddSteamInfo(mods)
	sortedMods, err := SortMods(mods)
	if err != nil {
		return err
	}

	fmt.Println(ExportMarkdown(sortedMods, config))
	return nil
}
func CmdToddsClean(context.Context, *cli.Command) error {
	config := LoadConfig()
	return ToddsClean(config)
}
func CmdToddsEncode(context.Context, *cli.Command) error {
	config := LoadConfig()
	return ToddsEncode(config)
}
func CmdSteam(context.Context, *cli.Command) error {
	config := LoadConfig()
	filename := "listpaths.tsv"

	mods, err := GetModsFromPath(filename, config)
	if err != nil {
		return err
	}
	LinkMods(mods)
	AddSteamInfo(mods)
	return nil
}

func CmdInstall(ctx context.Context, cmd *cli.Command) error {
	config := LoadConfig()

	args := cmd.Args().Slice()
	steamIDs := make([]SteamID, 0, len(args))
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid SteamID %q: %w", arg, err)
		}
		steamIDs = append(steamIDs, SteamID(id))
	}

	return SteamCMDInstall(config, steamIDs)
}

func CmdCheck(context.Context, *cli.Command) error {
	config := LoadConfig()
	mods := GetAllModsPath(config)

	for _, path := range mods {
		_, err := ParseAbout(path)
		if err != nil {
			fmt.Printf("Unsuccessful about parse on %s\n", filepath.Base(path))
		}
		// if len(about.ForceLoadAfterByVersion) > 0 || len(about.IncompatibleWithByVersion) > 0 || len(about.LoadAfterByVersion) > 0 || len(about.LoadBeforeByVersion) > 0 || len(about.ModDependenciesByVersion) > 0 {
		// 	fmt.Println(about)
		// }
	}

	return nil
}

func main() {
	commands := []*cli.Command{
		{
			Name:   "vanilla",
			Usage:  "Load vanilla Rimworld with all expansions",
			Action: CmdVanilla,
		},
		{
			Name:   "check",
			Usage:  "Check mods in SteamModSrc for errors",
			Action: CmdCheck,
		}, {
			Name:   "tsv",
			Usage:  "output mods in TSV for use elswhere",
			Action: CmdTsv,
		}, {
			Name:   "load",
			Usage:  "output mods in TSV for use elswhere",
			Action: CmdLoad,
		}, {
			Name:   "markdown",
			Usage:  "markdown export",
			Action: CmdMarkdown,
		}, {
			Name:   "encode_clean",
			Usage:  "Todds encode clean",
			Action: CmdToddsClean,
		}, {
			Name:   "encode",
			Usage:  "Todds encode",
			Action: CmdToddsEncode,
		}, {
			Name:   "install",
			Usage:  "SteamCMD install",
			Action: CmdInstall,
		},
	}

	cmd := &cli.Command{Commands: commands}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
