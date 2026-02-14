# platforms

Translate platform identifier strings across package ecosystems.

An ARM64 Mac is `darwin/arm64` to Go, `darwin-arm64` to Node, `aarch64-apple-darwin` to Rust, `arm64-darwin` to RubyGems, and `macosx_11_0_arm64` to Python. This module provides a shared mapping between all of them.

```go
import "github.com/git-pkgs/platforms"

// Parse an ecosystem-specific string into a canonical Platform
p, _ := platforms.Parse(platforms.Go, "darwin/arm64")
// p.Arch == "aarch64", p.OS == "darwin"

// Format a Platform for a different ecosystem
s, _ := platforms.Format(platforms.Rust, p)
// s == "aarch64-apple-darwin"

// Or translate directly
s, _ = platforms.Translate(platforms.Go, platforms.RubyGems, "darwin/arm64")
// s == "arm64-darwin"

// Normalize to the preferred form
s, _ = platforms.Normalize(platforms.Python, "linux_x86_64")
// s == "manylinux_2_17_x86_64"
```

## Supported ecosystems

| Ecosystem | Example | Format |
|---|---|---|
| `go` | `linux/amd64` | `os/arch` |
| `node` | `darwin-arm64` | `os-arch` |
| `rust` | `x86_64-unknown-linux-gnu` | `arch-vendor-os-abi` |
| `rubygems` | `arm64-darwin` | `arch-os` |
| `python` | `manylinux_2_17_x86_64` | `tag_version_arch` |
| `debian` | `x86_64-linux-gnu` | `arch-os-abi` |
| `llvm` | `aarch64-apple-darwin` | `arch-vendor-os` |
| `nuget` | `linux-x64` | `os-arch` |
| `vcpkg` | `x64-linux` | `arch-os` |
| `conan` | `Linux/armv8` | `os/arch` (settings) |
| `homebrew` | `arm64_sonoma` | `arch_codename` |
| `swift` | `aarch64-apple-darwin` | LLVM triples |
| `kotlin` | `linuxX64` | `osArch` (camelCase) |
| `maven` | `linux-aarch_64` | `os-arch` |

## How it works

The mapping data lives in JSON files under `data/`. `arches.json` and `oses.json` map canonical names to per-ecosystem aliases. `platforms.json` has pre-computed full strings for common platforms that can't be composed mechanically (like RubyGems using `arm64` on macOS but `aarch64` on Linux).

Parsing tries the pre-computed index first, then falls back to decomposing the string using ecosystem-specific rules and resolving components through the alias tables.

See [SPEC.md](SPEC.md) for the full specification.
