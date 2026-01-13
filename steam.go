package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type SteamResponseWrapper struct {
	Response SteamResponse `json:"response"`
}

type SteamResponse struct {
	Result               int         `json:"result"`
	ResultCount          int         `json:"resultcount"`
	PublishedFileDetails []SteamInfo `json:"publishedfiledetails"`
}

type SteamInfo struct {
	PublishedFileID       string `json:"publishedfileid"`
	Result                int    `json:"result"`
	Creator               string `json:"creator"`
	CreatorAppID          int    `json:"creator_app_id"`
	ConsumerAppID         int    `json:"consumer_app_id"`
	Filename              string `json:"filename"`
	FileSize              string `json:"file_size"`
	FileURL               string `json:"file_url"`
	HContentFile          string `json:"hcontent_file"`
	PreviewURL            string `json:"preview_url"`
	HContentPreview       string `json:"hcontent_preview"`
	Title                 string `json:"title"`
	Description           string `json:"description"`
	TimeCreated           int64  `json:"time_created"`
	TimeUpdated           int64  `json:"time_updated"`
	Visibility            int    `json:"visibility"`
	Banned                int    `json:"banned"`
	BanReason             string `json:"ban_reason"`
	Subscriptions         int    `json:"subscriptions"`
	Favorited             int    `json:"favorited"`
	LifetimeSubscriptions int    `json:"lifetime_subscriptions"`
	LifetimeFavorited     int    `json:"lifetime_favorited"`
	Views                 int    `json:"views"`
	Tags                  []Tag  `json:"tags"`
}

type Tag struct {
	Tag string `json:"tag"`
}

type SteamID int

const steamCacheFile = "steam_cache.json"
const rimworldAppID = "294100"

var ErrSteamCMDMIA = errors.New("SteamCMD did not create the mod folder after install")

func SteamCMDInstall(config Config, ids []SteamID) error {
	args := []string{}
	args = append(args, "+logon", "anonymous")

	for _, id := range ids {
		args = append(args, "+workshop_download_item", rimworldAppID, strconv.Itoa(int(id)))
	}
	args = append(args, "+exit")
	fmt.Println(args)

	cmd := exec.Command("steamcmd", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return (fmt.Errorf("steamcmd exited with code %d\n", exitErr.ExitCode()))
		} else {
			return (fmt.Errorf("failed to run steamcmd: %v\n", err))
		}
	}

	succeeded := true
	for _, id := range ids {
		modPath := filepath.Join(config.SteamModSrc, strconv.Itoa(int(id)))
		if _, err := os.Stat(modPath); errors.Is(err, os.ErrNotExist) {
			succeeded = false
			fmt.Printf("After running SteamCMD, mod %d could not be found", id)
		}
	}

	if succeeded {
		return nil
	}

	return ErrSteamCMDMIA
}

func loadSteamCache() (map[SteamID]SteamInfo, error) {
	cache := make(map[SteamID]SteamInfo)

	f, err := os.Open(steamCacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cache); err != nil {
		return nil, err
	}

	return cache, nil
}

func saveSteamCache(cache map[SteamID]SteamInfo) error {
	tmp := steamCacheFile + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(cache); err != nil {
		return err
	}

	return os.Rename(tmp, steamCacheFile)
}

func GetSteamInfo(ids []SteamID) (map[SteamID]SteamInfo, error) {
	values := url.Values{}
	values.Set("itemcount", strconv.Itoa(len(ids)))

	for i, id := range ids {
		values.Set(fmt.Sprintf("publishedfileids[%d]", i), strconv.Itoa(int(id)))
	}

	resp, err := http.PostForm(
		"https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/",
		values,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var wrapper SteamResponseWrapper
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, err
	}

	result := make(map[SteamID]SteamInfo)
	for _, info := range wrapper.Response.PublishedFileDetails {
		id, err := strconv.Atoi(info.PublishedFileID)
		if err != nil {
			continue
		}
		result[SteamID(id)] = info
	}

	return result, nil
}

func GetSteamInfoCached(ids []SteamID) (map[SteamID]SteamInfo, error) {
	cache, err := loadSteamCache()
	if err != nil {
		return nil, err
	}

	var missing []SteamID
	for _, id := range ids {
		if _, ok := cache[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		fetched, err := GetSteamInfo(missing)
		if err != nil {
			return nil, err
		}

		maps.Copy(cache, fetched)

		if err := saveSteamCache(cache); err != nil {
			return nil, err
		}
	}

	result := make(map[SteamID]SteamInfo)
	for _, id := range ids {
		if info, ok := cache[id]; ok {
			result[SteamID(id)] = info
		}
	}

	return result, nil
}

func AddSteamInfo(mods []*Mod) {
	modsBySteamAppId := map[SteamID]*Mod{}
	ids := []SteamID{}
	for _, mod := range mods {
		if mod.Source == ModSourceSteam {
			appId := mod.GetPublishedAppID()
			ids = append(ids, SteamID(appId))
			modsBySteamAppId[appId] = mod
		}
	}
	modSteamInfo, err := GetSteamInfoCached(ids)
	if err != nil {
		return
	}

	for id, info := range modSteamInfo {
		modsBySteamAppId[id].SteamInfo = &info
	}
}
