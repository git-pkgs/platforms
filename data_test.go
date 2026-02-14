package platforms

import (
	"encoding/json"
	"testing"
)

func TestDataLoads(t *testing.T) {
	idx, err := loadData()
	if err != nil {
		t.Fatalf("loadData: %v", err)
	}
	if len(idx.arches) == 0 {
		t.Fatal("no arches loaded")
	}
	if len(idx.oses) == 0 {
		t.Fatal("no oses loaded")
	}
	if len(idx.platFile.Platforms) == 0 {
		t.Fatal("no platforms loaded")
	}
}

func TestArchesJSONValid(t *testing.T) {
	var data archData
	if err := json.Unmarshal(archesJSON, &data); err != nil {
		t.Fatalf("arches.json: %v", err)
	}

	ecos := []string{"go", "node", "rust", "rubygems", "python", "debian", "llvm", "nuget", "vcpkg", "conan", "homebrew", "swift", "kotlin", "maven"}
	for arch, ecoMap := range data {
		for _, eco := range ecos {
			raw, ok := ecoMap[eco]
			if !ok {
				t.Errorf("arch %s missing ecosystem %s", arch, eco)
			}
			// Value should be null, a string, or array of strings.
			if string(raw) == "null" {
				continue
			}
			var s string
			var arr []string
			if json.Unmarshal(raw, &s) != nil && json.Unmarshal(raw, &arr) != nil {
				t.Errorf("arch %s, eco %s: value is neither string nor array: %s", arch, eco, string(raw))
			}
		}
	}
}

func TestOsesJSONValid(t *testing.T) {
	var data osData
	if err := json.Unmarshal(osesJSON, &data); err != nil {
		t.Fatalf("oses.json: %v", err)
	}

	ecos := []string{"go", "node", "rust", "rubygems", "python", "debian", "llvm", "nuget", "vcpkg", "conan", "homebrew", "swift", "kotlin", "maven"}
	for osName, ecoMap := range data {
		for _, eco := range ecos {
			raw, ok := ecoMap[eco]
			if !ok {
				t.Errorf("os %s missing ecosystem %s", osName, eco)
			}
			if string(raw) == "null" {
				continue
			}
			var s string
			var arr []string
			if json.Unmarshal(raw, &s) != nil && json.Unmarshal(raw, &arr) != nil {
				t.Errorf("os %s, eco %s: value is neither string nor array: %s", osName, eco, string(raw))
			}
		}
	}
}

func TestPlatformsJSONConsistency(t *testing.T) {
	idx, err := loadData()
	if err != nil {
		t.Fatalf("loadData: %v", err)
	}

	for i, p := range idx.platFile.Platforms {
		// Every arch in platforms.json should exist in arches.json.
		if _, ok := idx.arches[p.Arch]; !ok {
			t.Errorf("platform %d: arch %q not found in arches.json", i, p.Arch)
		}
		// Every OS should exist in oses.json.
		if _, ok := idx.oses[p.OS]; !ok {
			t.Errorf("platform %d: os %q not found in oses.json", i, p.OS)
		}
	}
}

func TestNoDuplicateStringsWithinEcosystem(t *testing.T) {
	idx, err := loadData()
	if err != nil {
		t.Fatalf("loadData: %v", err)
	}

	ecos := []string{"go", "node", "rust", "rubygems", "python", "debian", "llvm", "nuget", "vcpkg", "conan", "homebrew", "swift", "kotlin", "maven"}
	for _, eco := range ecos {
		seen := make(map[string]int) // string -> first platform index
		for i, p := range idx.platFile.Platforms {
			raw, ok := p.Strings[eco]
			if !ok {
				continue
			}
			_, all := resolveMapping(raw)
			for _, s := range all {
				if prev, exists := seen[s]; exists {
					t.Errorf("ecosystem %s: duplicate string %q in platform %d and %d", eco, s, prev, i)
				}
				seen[s] = i
			}
		}
	}
}

func TestReverseIndicesBuilt(t *testing.T) {
	idx, err := loadData()
	if err != nil {
		t.Fatalf("loadData: %v", err)
	}

	// Spot check: Go "amd64" should resolve to "x86_64".
	if got := resolveArch(idx, Go, "amd64"); got != "x86_64" {
		t.Errorf("resolveArch(Go, amd64) = %q, want x86_64", got)
	}

	// Node "win32" should resolve to "windows".
	if got := resolveOS(idx, Node, "win32"); got != "windows" {
		t.Errorf("resolveOS(Node, win32) = %q, want windows", got)
	}
}
