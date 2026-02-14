package platforms

import (
	"testing"
)

func TestEcosystems(t *testing.T) {
	ecos := Ecosystems()
	if len(ecos) != 14 {
		t.Fatalf("expected 14 ecosystems, got %d", len(ecos))
	}

	// Should be sorted.
	for i := 1; i < len(ecos); i++ {
		if ecos[i] < ecos[i-1] {
			t.Errorf("ecosystems not sorted: %s before %s", ecos[i-1], ecos[i])
		}
	}

	// Check all expected ecosystems are present.
	want := map[Ecosystem]bool{Go: true, Node: true, Rust: true, RubyGems: true, Python: true, Debian: true, LLVM: true, NuGet: true, Vcpkg: true, Conan: true, Homebrew: true, Swift: true, Kotlin: true, Maven: true}
	for _, e := range ecos {
		if !want[e] {
			t.Errorf("unexpected ecosystem: %s", e)
		}
		delete(want, e)
	}
	for e := range want {
		t.Errorf("missing ecosystem: %s", e)
	}
}

func TestPlatformZeroValue(t *testing.T) {
	var p Platform
	if p.Arch != "" || p.OS != "" || p.Vendor != "" || p.ABI != "" {
		t.Errorf("zero value should have empty fields")
	}
}

func TestErrorMessages(t *testing.T) {
	e1 := &ErrUnknownEcosystem{Ecosystem: "fake"}
	if e1.Error() != "unknown ecosystem: fake" {
		t.Errorf("unexpected error message: %s", e1.Error())
	}

	e2 := &ErrUnknownPlatform{Ecosystem: Go, Input: "bogus"}
	if e2.Error() != "unknown platform string for go: bogus" {
		t.Errorf("unexpected error message: %s", e2.Error())
	}

	e3 := &ErrNoMapping{Ecosystem: Debian, Platform: Platform{Arch: "x86_64", OS: "darwin"}}
	if e3.Error() != "no mapping for x86_64/darwin in debian" {
		t.Errorf("unexpected error message: %s", e3.Error())
	}
}
