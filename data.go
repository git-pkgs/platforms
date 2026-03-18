package platforms

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed data/arches.json
var archesJSON []byte

//go:embed data/oses.json
var osesJSON []byte

//go:embed data/platforms.json
var platformsJSON []byte

const jsonNull = "null"

// rawMapping is a JSON value that can be a string, array of strings, or null.
type rawMapping = json.RawMessage

// archData maps canonical arch -> ecosystem -> name(s).
type archData map[string]map[string]rawMapping

// osData maps canonical OS -> ecosystem -> name(s).
type osData map[string]map[string]rawMapping

type platformEntry struct {
	Arch    string                   `json:"arch"`
	OS      string                   `json:"os"`
	Vendor  string                   `json:"vendor"`
	ABI     string                   `json:"abi"`
	Strings map[string]rawMapping    `json:"strings"`
}

type platformsFile struct {
	Platforms []platformEntry `json:"platforms"`
}

// indices built at load time for fast lookups.
type indices struct {
	arches   archData
	oses     osData
	platFile platformsFile

	// archReverse maps ecosystem -> eco-specific name -> canonical arch.
	archReverse map[string]map[string]string
	// osReverse maps ecosystem -> eco-specific name -> canonical OS.
	osReverse map[string]map[string]string

	// platByString maps ecosystem -> platform string -> index into platFile.Platforms.
	platByString map[string]map[string]int
	// platByKey maps "arch/os/vendor/abi" -> index into platFile.Platforms.
	platByKey map[string]int
}

var (
	loadOnce sync.Once
	loaded   *indices
	loadErr  error
)

func loadData() (*indices, error) {
	loadOnce.Do(func() {
		idx := &indices{}

		if err := json.Unmarshal(archesJSON, &idx.arches); err != nil {
			loadErr = err
			return
		}
		if err := json.Unmarshal(osesJSON, &idx.oses); err != nil {
			loadErr = err
			return
		}
		if err := json.Unmarshal(platformsJSON, &idx.platFile); err != nil {
			loadErr = err
			return
		}

		idx.archReverse = buildReverse(idx.arches)
		idx.osReverse = buildReverse(idx.oses)
		idx.platByString = buildPlatStringIndex(idx.platFile)
		idx.platByKey = buildPlatKeyIndex(idx.platFile)

		loaded = idx
	})
	return loaded, loadErr
}

// resolveMapping extracts the string value(s) from a raw JSON mapping.
// Returns the preferred (first) value and all values.
func resolveMapping(raw rawMapping) (preferred string, all []string) {
	if raw == nil || string(raw) == jsonNull {
		return "", nil
	}

	// Try as a plain string first.
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, []string{s}
	}

	// Try as an array of strings.
	var arr []string
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		return arr[0], arr
	}

	return "", nil
}

// buildReverse builds a reverse lookup from ecosystem-specific name to
// canonical name for a set of mappings (arches or oses).
func buildReverse(data map[string]map[string]rawMapping) map[string]map[string]string {
	rev := make(map[string]map[string]string)
	for canonical, ecoMap := range data {
		for eco, raw := range ecoMap {
			if rev[eco] == nil {
				rev[eco] = make(map[string]string)
			}
			_, all := resolveMapping(raw)
			for _, name := range all {
				rev[eco][strings.ToLower(name)] = canonical
			}
		}
	}
	return rev
}

func buildPlatStringIndex(pf platformsFile) map[string]map[string]int {
	idx := make(map[string]map[string]int)
	for i, p := range pf.Platforms {
		for eco, raw := range p.Strings {
			if idx[eco] == nil {
				idx[eco] = make(map[string]int)
			}
			_, all := resolveMapping(raw)
			for _, s := range all {
				idx[eco][strings.ToLower(s)] = i
			}
		}
	}
	return idx
}

func platKey(arch, os, vendor, abi string) string {
	return arch + "/" + os + "/" + vendor + "/" + abi
}

func buildPlatKeyIndex(pf platformsFile) map[string]int {
	idx := make(map[string]int)
	for i, p := range pf.Platforms {
		idx[platKey(p.Arch, p.OS, p.Vendor, p.ABI)] = i
	}
	return idx
}

// resolveArch returns the canonical arch for an ecosystem-specific name.
func resolveArch(idx *indices, eco Ecosystem, name string) string {
	if m, ok := idx.archReverse[string(eco)]; ok {
		if canonical, ok := m[strings.ToLower(name)]; ok {
			return canonical
		}
	}
	return ""
}

// resolveOS returns the canonical OS for an ecosystem-specific name.
func resolveOS(idx *indices, eco Ecosystem, name string) string {
	if m, ok := idx.osReverse[string(eco)]; ok {
		if canonical, ok := m[strings.ToLower(name)]; ok {
			return canonical
		}
	}
	return ""
}

// lookupArch returns the preferred ecosystem-specific name for a canonical arch.
func lookupArch(idx *indices, eco Ecosystem, canonical string) string {
	if ecoMap, ok := idx.arches[canonical]; ok {
		if raw, ok := ecoMap[string(eco)]; ok {
			pref, _ := resolveMapping(raw)
			return pref
		}
	}
	return ""
}

// lookupOS returns the preferred ecosystem-specific name for a canonical OS.
func lookupOS(idx *indices, eco Ecosystem, canonical string) string {
	if ecoMap, ok := idx.oses[canonical]; ok {
		if raw, ok := ecoMap[string(eco)]; ok {
			pref, _ := resolveMapping(raw)
			return pref
		}
	}
	return ""
}
