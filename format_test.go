package platforms

import (
	"errors"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		eco  Ecosystem
		plat Platform
		want string
	}{
		// x86_64 Linux GNU
		{Go, Platform{Arch: "x86_64", OS: "linux"}, "linux/amd64"},
		{Node, Platform{Arch: "x86_64", OS: "linux"}, "linux-x64"},
		{Rust, Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "gnu"}, "x86_64-unknown-linux-gnu"},
		{RubyGems, Platform{Arch: "x86_64", OS: "linux"}, "x86_64-linux"},
		{Python, Platform{Arch: "x86_64", OS: "linux", ABI: "gnu"}, "manylinux_2_17_x86_64"},
		{Python, Platform{Arch: "x86_64", OS: "linux", ABI: "gnu", LibCVersion: "2.28"}, "manylinux_2_28_x86_64"},
		{Debian, Platform{Arch: "x86_64", OS: "linux", ABI: "gnu"}, "x86_64-linux-gnu"},
		{LLVM, Platform{Arch: "x86_64", OS: "linux", Vendor: "pc", ABI: "gnu"}, "x86_64-pc-linux-gnu"},

		// ARM64 macOS
		{Go, Platform{Arch: "aarch64", OS: "darwin"}, "darwin/arm64"},
		{Node, Platform{Arch: "aarch64", OS: "darwin"}, "darwin-arm64"},
		{Rust, Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}, "aarch64-apple-darwin"},
		{RubyGems, Platform{Arch: "aarch64", OS: "darwin"}, "arm64-darwin"},
		{Python, Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}, "macosx_11_0_arm64"},
		{Python, Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple", OSVersion: "14.0"}, "macosx_14_0_arm64"},
		{LLVM, Platform{Arch: "aarch64", OS: "darwin", Vendor: "apple"}, "aarch64-apple-darwin"},

		// x86_64 Windows MSVC
		{Go, Platform{Arch: "x86_64", OS: "windows"}, "windows/amd64"},
		{Node, Platform{Arch: "x86_64", OS: "windows"}, "win32-x64"},
		{Rust, Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}, "x86_64-pc-windows-msvc"},
		{Python, Platform{Arch: "x86_64", OS: "windows"}, "win_amd64"},
		{LLVM, Platform{Arch: "x86_64", OS: "windows", Vendor: "pc", ABI: "msvc"}, "x86_64-pc-windows-msvc"},

		// ARM64 Linux GNU
		{Go, Platform{Arch: "aarch64", OS: "linux"}, "linux/arm64"},
		{Rust, Platform{Arch: "aarch64", OS: "linux", Vendor: "unknown", ABI: "gnu"}, "aarch64-unknown-linux-gnu"},
		{Debian, Platform{Arch: "aarch64", OS: "linux", ABI: "gnu"}, "aarch64-linux-gnu"},

		// Musl
		{Rust, Platform{Arch: "x86_64", OS: "linux", Vendor: "unknown", ABI: "musl"}, "x86_64-unknown-linux-musl"},
		{RubyGems, Platform{Arch: "x86_64", OS: "linux", ABI: "musl"}, "x86_64-linux-musl"},
		{Python, Platform{Arch: "x86_64", OS: "linux", ABI: "musl"}, "musllinux_1_1_x86_64"},

		// i686 Windows
		{Python, Platform{Arch: "i686", OS: "windows"}, "win32"},

		// FreeBSD
		{Go, Platform{Arch: "x86_64", OS: "freebsd"}, "freebsd/amd64"},
		{Rust, Platform{Arch: "x86_64", OS: "freebsd"}, "x86_64-unknown-freebsd"},

		// Compose fallback for unlisted platforms
		{Go, Platform{Arch: "mips64le", OS: "linux"}, "linux/mips64le"},
		{Rust, Platform{Arch: "s390x", OS: "linux", Vendor: "unknown", ABI: "gnu"}, "s390x-unknown-linux-gnu"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eco)+"/"+tt.want, func(t *testing.T) {
			got, err := Format(tt.eco, tt.plat)
			if err != nil {
				t.Fatalf("Format(%s, %+v) error: %v", tt.eco, tt.plat, err)
			}
			if got != tt.want {
				t.Errorf("Format(%s, %+v) = %q, want %q", tt.eco, tt.plat, got, tt.want)
			}
		})
	}
}

func TestFormatNoMapping(t *testing.T) {
	// Darwin is not supported by Debian.
	_, err := Format(Debian, Platform{Arch: "x86_64", OS: "darwin"})
	var e *ErrNoMapping
	if !errors.As(err, &e) {
		t.Fatalf("expected ErrNoMapping, got %v", err)
	}
}

func TestFormatUnknownEcosystem(t *testing.T) {
	_, err := Format("bogus", Platform{Arch: "x86_64", OS: "linux"})
	var e *ErrUnknownEcosystem
	if !errors.As(err, &e) {
		t.Fatalf("expected ErrUnknownEcosystem, got %v", err)
	}
}
