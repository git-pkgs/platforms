package platforms

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		eco  Ecosystem
		in   string
		want Platform
	}{
		// Blog post comparison table: 64-bit x86 Linux
		// Go/Node hit the pre-computed gnu entry (vendor=unknown, abi=gnu)
		{Go, "linux/amd64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Node, "linux-x64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Rust, "x86_64-unknown-linux-gnu", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{RubyGems, "x86_64-linux", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Python, "manylinux_2_17_x86_64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu", LibCVersion: "2.17"}},
		{Debian, "x86_64-linux-gnu", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		// LLVM pre-computed entry has vendor=unknown (canonical), not pc
		{LLVM, "x86_64-pc-linux-gnu", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// Blog post comparison table: ARM64 macOS
		{Go, "darwin/arm64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Node, "darwin-arm64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Rust, "aarch64-apple-darwin", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{RubyGems, "arm64-darwin", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Python, "macosx_11_0_arm64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple", OSVersion: "11.0"}},
		{LLVM, "aarch64-apple-darwin", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},

		// Blog post comparison table: 64-bit x86 Windows
		{Go, "windows/amd64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},
		{Node, "win32-x64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},
		{Rust, "x86_64-pc-windows-msvc", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},
		{RubyGems, "x64-mingw-ucrt", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},
		{Python, "win_amd64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},
		{LLVM, "x86_64-pc-windows-msvc", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// Blog post comparison table: ARM64 Linux
		{Go, "linux/arm64", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Node, "linux-arm64", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Rust, "aarch64-unknown-linux-gnu", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{RubyGems, "aarch64-linux", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Python, "manylinux_2_17_aarch64", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu", LibCVersion: "2.17"}},
		{Debian, "aarch64-linux-gnu", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		// LLVM pre-computed entry has vendor=unknown
		{LLVM, "aarch64-pc-linux-gnu", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// Musl Linux
		{Rust, "x86_64-unknown-linux-musl", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "musl"}},
		{RubyGems, "x86_64-linux-musl", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "musl"}},
		{Python, "musllinux_1_1_x86_64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "musl", LibCVersion: "1.1"}},

		// Python aliases (hit pre-computed gnu entry)
		{Python, "linux_x86_64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Python, "linux_aarch64", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// Windows 32-bit
		{Python, "win32", Platform{Arch: "i686", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// macOS x86_64
		{Rust, "x86_64-apple-darwin", Platform{Arch: "x86_64", OS: "darwin", Vendor: "apple"}},
		{Python, "macosx_10_9_x86_64", Platform{Arch: "x86_64", OS: "darwin", Vendor: "apple", OSVersion: "10.9"}},

		// RubyGems arm64-darwin vs aarch64-linux
		{RubyGems, "arm64-darwin", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{RubyGems, "aarch64-linux", Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// FreeBSD (pre-computed, vendor=unknown)
		{Go, "freebsd/amd64", Platform{Arch: "x86_64", OS: "freebsd", Vendor: "unknown"}},

		// i686 Linux via Debian (pre-computed, vendor=unknown)
		{Debian, "i386-linux-gnu", Platform{Arch: "i686", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// Decomposition fallback for Go (no pre-computed entry)
		{Go, "linux/mips64le", Platform{Arch: "mips64le", OS: "linux"}},
		{Go, "plan9/amd64", Platform{Arch: "x86_64", OS: "plan9"}},

		// RubyGems mingw32 -> mingw ABI entry
		{RubyGems, "x64-mingw32", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "mingw"}},

		// NuGet
		{NuGet, "linux-x64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{NuGet, "linux-musl-x64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "musl"}},
		{NuGet, "osx-arm64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{NuGet, "win-x64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// vcpkg
		{Vcpkg, "x64-linux", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Vcpkg, "arm64-osx", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Vcpkg, "x64-windows", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// Swift (uses LLVM triples)
		{Swift, "aarch64-apple-darwin", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Swift, "x86_64-unknown-linux-gnu", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},

		// Kotlin
		{Kotlin, "linuxX64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Kotlin, "macosArm64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Kotlin, "mingwX64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// Maven
		{Maven, "linux-x86_64", Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}},
		{Maven, "osx-aarch_64", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
		{Maven, "windows-x86_64", Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}},

		// Homebrew
		{Homebrew, "arm64_sonoma", Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.eco)+"/"+tt.in, func(t *testing.T) {
			got, err := Parse(tt.eco, tt.in)
			if err != nil {
				t.Fatalf("Parse(%s, %q) error: %v", tt.eco, tt.in, err)
			}
			if got.Arch != tt.want.Arch {
				t.Errorf("Arch = %q, want %q", got.Arch, tt.want.Arch)
			}
			if got.OS != tt.want.OS {
				t.Errorf("OS = %q, want %q", got.OS, tt.want.OS)
			}
			if got.Vendor != tt.want.Vendor {
				t.Errorf("Vendor = %q, want %q", got.Vendor, tt.want.Vendor)
			}
			if got.ABI != tt.want.ABI {
				t.Errorf("ABI = %q, want %q", got.ABI, tt.want.ABI)
			}
			if got.OSVersion != tt.want.OSVersion {
				t.Errorf("OSVersion = %q, want %q", got.OSVersion, tt.want.OSVersion)
			}
			if got.LibCVersion != tt.want.LibCVersion {
				t.Errorf("LibCVersion = %q, want %q", got.LibCVersion, tt.want.LibCVersion)
			}
		})
	}
}

func TestParseUnknownEcosystem(t *testing.T) {
	_, err := Parse("bogus", "whatever")
	var e *ErrUnknownEcosystem
	if !errors.As(err, &e) {
		t.Fatalf("expected ErrUnknownEcosystem, got %v", err)
	}
}

func TestParseUnknownPlatform(t *testing.T) {
	_, err := Parse(Go, "nope/nada")
	var e *ErrUnknownPlatform
	if !errors.As(err, &e) {
		t.Fatalf("expected ErrUnknownPlatform, got %v", err)
	}
}
