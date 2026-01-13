package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

func LinkMods(mods []*Mod) {
	modsByPid := map[PackageID]*Mod{}
	for _, mod := range mods {
		modsByPid[mod.PackageID] = mod
	}
	for _, mod := range mods {
		for _, pid := range mod.LoadAfterFull() {
			if target, ok := modsByPid[pid]; ok {
				if !slices.Contains(mod.LoadAfter, target) {
					mod.LoadAfter = append(mod.LoadAfter, target)
				}
			}
		}
		for _, pid := range mod.LoadBeforeFull() {
			if target, ok := modsByPid[pid]; ok {
				if !slices.Contains(target.LoadAfter, mod) {
					target.LoadAfter = append(target.LoadAfter, mod)
				}
			}
		}
	}
}

func CheckDeps(mods []*Mod, config Config) error {
	version := GetRimworldVersion(config)[:3]
	modsByPid := map[PackageID]*Mod{}
	for _, mod := range mods {
		modsByPid[mod.PackageID] = mod
	}
	missing := false
ModLoop:
	for _, mod := range mods {
		supported := false
		for _, modsupport := range mod.About.SupportedVersions {
			if modsupport == version {
				supported = true
			}
		}
		if !supported {
			if string(mod.PackageID) != "ludeon.rimworld" {
				fmt.Printf("Mod has wrong version %s\n", mod)
			}
		}
		for _, deps := range mod.Deps {
			for _, dep := range deps {
				if _, ok := modsByPid[dep]; ok {
					continue ModLoop
				}
				fmt.Printf("Missing required dependency %s\n", dep)
				missing = true
			}
		}
	}
	if missing {
		return fmt.Errorf("Missing dependencies")
	} else {
		return nil
	}
}

func findCycle(graph map[*Mod][]*Mod) []*Mod {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[*Mod]int)
	parent := make(map[*Mod]*Mod)

	var cycle []*Mod
	var dfs func(*Mod) bool

	dfs = func(u *Mod) bool {
		state[u] = visiting

		for _, v := range graph[u] {
			if state[v] == unvisited {
				parent[v] = u
				if dfs(v) {
					return true
				}
			} else if state[v] == visiting {
				cycle = append(cycle, v)
				for x := u; x != v; x = parent[x] {
					cycle = append(cycle, x)
				}
				cycle = append(cycle, v)
				return true
			}
		}

		state[u] = visited
		return false
	}

	for node := range graph {
		if state[node] == unvisited {
			if dfs(node) {
				break
			}
		}
	}

	for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
		cycle[i], cycle[j] = cycle[j], cycle[i]
	}

	return cycle
}
func SortMods(mods []*Mod) ([]*Mod, error) {
	indegree := make(map[*Mod]int)
	graph := make(map[*Mod][]*Mod)

	for _, mod := range mods {
		if _, ok := indegree[mod]; !ok {
			indegree[mod] = 0
		}
		for _, dep := range mod.LoadAfter {
			graph[dep] = append(graph[dep], mod)
			indegree[mod]++
			if _, ok := indegree[dep]; !ok {
				indegree[dep] = 0
			}
		}
	}

	var ready []*Mod
	for mod, deg := range indegree {
		if deg == 0 {
			ready = append(ready, mod)
		}
	}

	sort.Slice(ready, func(i, j int) bool {
		return ready[i].PackageID < ready[j].PackageID
	})

	var result []*Mod

	for len(ready) > 0 {
		mod := ready[0]
		ready = ready[1:]

		result = append(result, mod)

		for _, next := range graph[mod] {
			indegree[next]--
			if indegree[next] == 0 {
				ready = append(ready, next)
			}
		}

		sort.Slice(ready, func(i, j int) bool {
			return ready[i].PackageID < ready[j].PackageID
		})
	}

	if len(result) != len(indegree) {
		cycle := findCycle(graph)

		var names []string
		for _, m := range cycle {
			names = append(names, string(m.PackageID))
		}

		return nil, fmt.Errorf(
			"cycle detected in dependency graph: %s",
			strings.Join(names, " -> "),
		)
	}

	return result, nil
}
