// Package platforms translates platform identifier strings across package
// ecosystems. An ARM64 Mac is "darwin/arm64" to Go, "darwin-arm64" to Node,
// "aarch64-apple-darwin" to Rust, "arm64-darwin" to RubyGems, and
// "macosx_11_0_arm64" to Python. This package provides a shared mapping
// between all of them.
package platforms

import "sort"

// Ecosystem identifies a package ecosystem's platform naming convention.
type Ecosystem string

const (
	Go       Ecosystem = "go"
	Node     Ecosystem = "node"
	Rust     Ecosystem = "rust"
	RubyGems Ecosystem = "rubygems"
	Python   Ecosystem = "python"
	Debian   Ecosystem = "debian"
	LLVM     Ecosystem = "llvm"
)

var allEcosystems = []Ecosystem{Go, Node, Rust, RubyGems, Python, Debian, LLVM}

// Ecosystems returns all supported ecosystem identifiers in sorted order.
func Ecosystems() []Ecosystem {
	out := make([]Ecosystem, len(allEcosystems))
	copy(out, allEcosystems)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Platform represents a canonical platform with architecture, OS, and
// optional vendor, ABI, OS version, and libc version fields.
type Platform struct {
	Arch        string
	OS          string
	Vendor      string
	ABI         string
	OSVersion   string
	LibCVersion string
}

// ErrUnknownEcosystem is returned when an unrecognized ecosystem is provided.
type ErrUnknownEcosystem struct {
	Ecosystem string
}

func (e *ErrUnknownEcosystem) Error() string {
	return "unknown ecosystem: " + e.Ecosystem
}

// ErrUnknownPlatform is returned when a platform string cannot be parsed.
type ErrUnknownPlatform struct {
	Ecosystem Ecosystem
	Input     string
}

func (e *ErrUnknownPlatform) Error() string {
	return "unknown platform string for " + string(e.Ecosystem) + ": " + e.Input
}

// ErrNoMapping is returned when a platform cannot be formatted for an ecosystem.
type ErrNoMapping struct {
	Ecosystem Ecosystem
	Platform  Platform
}

func (e *ErrNoMapping) Error() string {
	return "no mapping for " + e.Platform.Arch + "/" + e.Platform.OS + " in " + string(e.Ecosystem)
}
