package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type About struct {
	PackageID         string          `xml:"packageId"`
	XMLName           xml.Name        `xml:"ModMetaData"`
	Name              string          `xml:"name"`
	ModVersion        string          `xml:"modVersion"`
	Url               string          `xml:"url"`
	Author            string          `xml:"author"`
	Authors           []string        `xml:"authors>li"`
	SupportedVersions []string        `xml:"supportedVersions>li"`
	LoadAfter         []string        `xml:"loadAfter>li"`
	LoadBefore        []string        `xml:"loadBefore>li"`
	ForceLoadAfter    []string        `xml:"forceLoadAfter>li"`
	ForceLoadBefore   []string        `xml:"forceLoadBefore>li"`
	ModDependencies   []ModDependency `xml:"modDependencies>li"`
	IncompatibleWith  []string        `xml:"incompatibleWith>li"`
	Description       string          `xml:"description"`

	// in my rather extensive modlist, these are used literally nowhere
	LoadAfterByVersion        []string        `xml:"loadAfterByVersion>li"`
	LoadBeforeByVersion       []string        `xml:"loadBeforeByVersion>li"`
	ForceLoadAfterByVersion   []string        `xml:"forceLoadAfterByVersion>li"`
	ModDependenciesByVersion  []ModDependency `xml:"modDependenciesByVersion>li"`
	IncompatibleWithByVersion []string        `xml:"incompatibleWithByVersion>li"`
}

type ModDependency struct {
	PackageID             string   `xml:"packageId"`
	DisplayName           string   `xml:"displayName"`
	SteamWorkshopURL      string   `xml:"steamWorkshopUrl"`
	DownloadURL           string   `xml:"downloadUrl"`
	AlternativePackageIds []string `xml:"alternativePackageIds>li"`
}

type ModSource int

const (
	ModSourceLocal ModSource = iota
	ModSourceSteam
	ModSourceGit
	ModSourceOfficial
)

type PackageID string

type Mod struct {
	Path      string
	PackageID PackageID
	Source    ModSource
	About     About
	LoadAfter []*Mod
	// inner list is "one of the following"
	Deps      [][]PackageID
	SteamInfo *SteamInfo
}

func (mod *Mod) TSVInfo() []string {
	res := []string{}
	res = append(res, string(mod.About.Name))
	res = append(res, string(mod.About.Url))
	res = append(res, string(mod.PackageID))
	res = append(res, string(mod.Path))
	res = append(res, string(strings.Join(mod.About.SupportedVersions, ", ")))
	return res
}

func (mod *Mod) String() string {
	return fmt.Sprintf("%s @ %s", string(mod.PackageID), string(mod.Path))
}

func (mod *Mod) LoadAfterFull() []PackageID {
	out := []PackageID{}
	for _, pid := range mod.About.LoadAfter {
		out = append(out, PackageID(strings.ToLower(pid)))
	}
	for _, pid := range mod.About.ForceLoadAfter {
		out = append(out, PackageID(strings.ToLower(pid)))
	}
	// for _, pid := range mod.About.LoadAfterByVersion {
	// 	out = append(out, PackageID(pid))
	// }
	if !slices.Contains(mod.LoadBeforeFull(), "ludeon.rimworld") && !(mod.Source == ModSourceOfficial) {
		out = append(out, PackageID("ludeon.rimworld"))
		out = append(out, PackageID("ludeon.rimworld.anomaly"))
		out = append(out, PackageID("ludeon.rimworld.odyssey"))
		out = append(out, PackageID("ludeon.rimworld.royalty"))
		out = append(out, PackageID("ludeon.rimworld.ideology"))
		out = append(out, PackageID("ludeon.rimworld.biotech"))
	}
	return out
}
func (mod *Mod) LoadBeforeFull() []PackageID {
	out := []PackageID{}
	for _, pid := range mod.About.LoadBefore {
		out = append(out, PackageID(strings.ToLower(pid)))
	}
	for _, pid := range mod.About.ForceLoadBefore {
		out = append(out, PackageID(strings.ToLower(pid)))
	}
	// for _, pid := range mod.About.LoadAfterByVersion {
	// 	out = append(out, PackageID(pid))
	// }
	return out
}
func (mod *Mod) BestSupportedVersion() string {
	modVersions := []float64{}
	for _, version := range mod.About.SupportedVersions {
		versionFloat, err := strconv.ParseFloat(version, 64)
		if err != nil {
			fmt.Printf("Error parsing version string %s", version)
		}
		modVersions = append(modVersions, versionFloat)
	}

	var bestVersion float64
	if len(modVersions) > 0 {
		bestVersion = slices.Max(modVersions)
	} else {
		bestVersion = 1.6
	}
	return fmt.Sprint(bestVersion)
}
func (mod *Mod) GetPublishedAppID() SteamID {
	content, err := os.ReadFile(filepath.Join(mod.Path, "About/PublishedFileId.txt"))
	if err != nil {
		return 0
	}
	idStr := strings.TrimRight(string(content), "\r\n")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid steam ID " + idStr)
		return 0
	}
	return SteamID(id)
}

var ErrInvalidAbout = errors.New("About parse failed")
var ErrDuplicatePID = errors.New("Duplicate PackageID")

func ParseAbout(path string) (*About, error) {
	aboutPath := filepath.Join(path, "About/About.xml")
	data, err := os.ReadFile(aboutPath)

	// horrible, I know
	if err != nil {
		data, err = os.ReadFile(filepath.Join(path, "About/about.xml"))
		if err != nil {
			data, err = os.ReadFile(filepath.Join(path, "about/About.xml"))
			if err != nil {
				data, err = os.ReadFile(filepath.Join(path, "about/about.xml"))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	var result About
	err = xml.Unmarshal(data, &result)
	if err != nil {
		return nil, ErrInvalidAbout
	}
	return &result, nil
}

func GetModSource(path string, config Config) ModSource {
	dir := filepath.Dir(path)
	if dir == config.SteamModSrc {
		return ModSourceSteam
	}
	if dir == filepath.Join(config.TargetDir, "Data") {
		return ModSourceOfficial
	}
	if dir == config.LocalModSrc {
		_, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil {
			return ModSourceGit
		}
	}
	return ModSourceLocal
}

func ParseMod(path string, config Config) (*Mod, error) {
	about, err := ParseAbout(path)
	if err != nil {
		return nil, err
	}

	deps := [][]PackageID{}
	for _, dep := range about.ModDependencies {
		allowed_deps := []PackageID{}
		allowed_deps = append(allowed_deps, PackageID(strings.ToLower(dep.PackageID)))
		for _, alternate := range dep.AlternativePackageIds {
			allowed_deps = append(allowed_deps, PackageID(strings.ToLower(alternate)))
		}
		deps = append(deps, allowed_deps)
	}

	return &Mod{
		Path:      path,
		Source:    GetModSource(path, config),
		PackageID: PackageID(strings.ToLower(about.PackageID)),
		About:     *about,
		Deps:      deps,
	}, nil
}

func GetAllModsPath(config Config) []string {
	steamModPaths, err := os.ReadDir(config.SteamModSrc)
	if err != nil {
		fmt.Println(err)
	}
	localModPaths, err := os.ReadDir(config.LocalModSrc)
	if err != nil {
		fmt.Println(err)
	}
	officialModPaths, err := os.ReadDir(filepath.Join(config.TargetDir, "Data"))
	if err != nil {
		fmt.Println(err)
	}

	paths := []string{}
	for _, entry := range steamModPaths {
		if !entry.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(config.SteamModSrc, entry.Name()))
	}
	for _, entry := range localModPaths {
		if !entry.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(config.LocalModSrc, entry.Name()))
	}
	for _, entry := range officialModPaths {
		if !entry.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(filepath.Join(config.TargetDir, "Data"), entry.Name()))
	}
	return paths
}

func GetAllMods(config Config) []*Mod {
	mods := []*Mod{}
	for _, modPath := range GetAllModsPath(config) {
		mod, err := ParseMod(modPath, config)
		if err != nil {
			continue
		}
		mods = append(mods, mod)
	}
	return mods
}
