package platforms

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		from, to Ecosystem
		in, want string
	}{
		// Go -> everything
		{Go, Rust, "linux/amd64", "x86_64-unknown-linux-gnu"},
		{Go, Node, "linux/amd64", "linux-x64"},
		{Go, RubyGems, "darwin/arm64", "arm64-darwin"},
		{Go, Python, "darwin/arm64", "macosx_11_0_arm64"},
		{Go, LLVM, "linux/arm64", "aarch64-pc-linux-gnu"},

		// Rust -> Go
		{Rust, Go, "x86_64-unknown-linux-gnu", "linux/amd64"},
		{Rust, Go, "aarch64-apple-darwin", "darwin/arm64"},
		{Rust, Go, "x86_64-pc-windows-msvc", "windows/amd64"},

		// Node -> Rust
		{Node, Rust, "linux-x64", "x86_64-unknown-linux-gnu"},
		{Node, Rust, "darwin-arm64", "aarch64-apple-darwin"},
		{Node, Rust, "win32-x64", "x86_64-pc-windows-msvc"},

		// Python -> Go
		{Python, Go, "manylinux_2_17_x86_64", "linux/amd64"},
		{Python, Go, "macosx_11_0_arm64", "darwin/arm64"},
		{Python, Go, "win_amd64", "windows/amd64"},

		// RubyGems -> Rust
		{RubyGems, Rust, "x86_64-linux", "x86_64-unknown-linux-gnu"},
		{RubyGems, Rust, "arm64-darwin", "aarch64-apple-darwin"},

		// Debian -> Go
		{Debian, Go, "x86_64-linux-gnu", "linux/amd64"},
		{Debian, Go, "aarch64-linux-gnu", "linux/arm64"},

		// LLVM -> RubyGems
		{LLVM, RubyGems, "x86_64-pc-linux-gnu", "x86_64-linux"},
		{LLVM, RubyGems, "aarch64-apple-darwin", "arm64-darwin"},

		// Cross-ecosystem with musl
		{Rust, RubyGems, "x86_64-unknown-linux-musl", "x86_64-linux-musl"},
		{RubyGems, Rust, "x86_64-linux-musl", "x86_64-unknown-linux-musl"},

		// NuGet -> Rust
		{NuGet, Rust, "linux-x64", "x86_64-unknown-linux-gnu"},
		{NuGet, Rust, "linux-musl-x64", "x86_64-unknown-linux-musl"},
		{NuGet, Rust, "osx-arm64", "aarch64-apple-darwin"},

		// Go -> vcpkg
		{Go, Vcpkg, "linux/amd64", "x64-linux"},
		{Go, Vcpkg, "darwin/arm64", "arm64-osx"},

		// Kotlin -> Go
		{Kotlin, Go, "linuxX64", "linux/amd64"},
		{Kotlin, Go, "macosArm64", "darwin/arm64"},

		// Maven -> Node
		{Maven, Node, "linux-x86_64", "linux-x64"},
		{Maven, Node, "osx-aarch_64", "darwin-arm64"},

		// Go -> NuGet
		{Go, NuGet, "linux/amd64", "linux-x64"},
		{Go, NuGet, "darwin/arm64", "osx-arm64"},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to)+"/"+tt.in, func(t *testing.T) {
			got, err := Translate(tt.from, tt.to, tt.in)
			if err != nil {
				t.Fatalf("Translate(%s, %s, %q) error: %v", tt.from, tt.to, tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Translate(%s, %s, %q) = %q, want %q", tt.from, tt.to, tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		eco      Ecosystem
		in, want string
	}{
		// RubyGems arm64 alias
		{RubyGems, "arm64-darwin", "arm64-darwin"},
		// Python linux alias
		{Python, "linux_x86_64", "manylinux_2_17_x86_64"},
		// Go identity
		{Go, "linux/amd64", "linux/amd64"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eco)+"/"+tt.in, func(t *testing.T) {
			got, err := Normalize(tt.eco, tt.in)
			if err != nil {
				t.Fatalf("Normalize(%s, %q) error: %v", tt.eco, tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Normalize(%s, %q) = %q, want %q", tt.eco, tt.in, got, tt.want)
			}
		})
	}
}

func TestTranslateRoundTrip(t *testing.T) {
	// Parse from A, format to B, parse from B, compare platforms.
	ecosystems := []Ecosystem{Go, Node, Rust, RubyGems}
	starts := map[Ecosystem]string{
		Go:       "linux/amd64",
		Node:     "linux-x64",
		Rust:     "x86_64-unknown-linux-gnu",
		RubyGems: "x86_64-linux",
	}

	for _, from := range ecosystems {
		for _, to := range ecosystems {
			t.Run(string(from)+"->"+string(to), func(t *testing.T) {
				translated, err := Translate(from, to, starts[from])
				if err != nil {
					t.Fatalf("Translate(%s, %s, %q) error: %v", from, to, starts[from], err)
				}

				// Parse the result back.
				p, err := Parse(to, translated)
				if err != nil {
					t.Fatalf("Parse(%s, %q) error: %v", to, translated, err)
				}

				// Core fields should match.
				if p.Arch != "x86_64" {
					t.Errorf("round-trip Arch = %q, want x86_64", p.Arch)
				}
				if p.OS != osLinux {
					t.Errorf("round-trip OS = %q, want linux", p.OS)
				}
			})
		}
	}
}
