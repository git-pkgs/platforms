# Platform String Specification

This document defines the canonical platform representation and the mapping rules used by the `platforms` module.

## Canonical fields

A platform has six fields. Only Arch and OS are required.

| Field | Example | Description |
|---|---|---|
| Arch | `x86_64` | CPU architecture, using kernel/vendor conventions |
| OS | `linux` | Operating system |
| Vendor | `apple` | Hardware vendor (optional) |
| ABI | `gnu` | ABI or C library variant (optional) |
| OSVersion | `11.0` | Minimum OS version (optional, used by Python macOS) |
| LibCVersion | `2.17` | Minimum libc version (optional, used by Python manylinux) |

## Canonical names

Architecture names follow `uname -m` output where possible:

- `x86_64` (not `amd64`, not `x64`)
- `aarch64` (not `arm64`)
- `i686` (not `x86`, not `i386`, not `386`)
- `arm`, `riscv64`, `s390x`, `ppc64`, `ppc64le`, `loong64`
- `mips`, `mips64`, `mipsle`, `mips64le`

OS names use lowercase kernel names:

- `linux`, `darwin`, `windows`, `freebsd`, `netbsd`, `openbsd`
- `android`, `ios`, `aix`, `solaris`, `dragonfly`, `illumos`, `plan9`

## Ecosystems

Fourteen ecosystems are supported: `go`, `node`, `rust`, `rubygems`, `python`, `debian`, `llvm`, `nuget`, `vcpkg`, `conan`, `homebrew`, `swift`, `kotlin`, `maven`.

## Data files

The mapping data lives in three JSON files under `data/`.

### arches.json

Maps each canonical arch to its name in each ecosystem. Values can be a string, an array of strings (first is preferred for formatting, all are recognized for parsing), or null (unsupported).

### oses.json

Same structure for OS names.

### platforms.json

Pre-computed full platform strings for common arch/os/vendor/abi combinations. Each entry maps a canonical platform to the complete string each ecosystem uses. This handles cases where the string can't be composed mechanically from arch and OS alone, like Python's `manylinux_2_17_x86_64` or RubyGems' use of `arm64` for darwin but `aarch64` for linux.

Only ecosystems that have a distinct string for the exact combination should appear in an entry. When an ecosystem can't distinguish between two entries (for example, Go produces `linux/amd64` for both gnu and musl), the string should appear only in the default (gnu) entry.

## Parse rules

1. Look up the input string in the `platforms.json` index for the given ecosystem
2. If not found, decompose using ecosystem-specific splitting:
   - **Go**: split on `/` as `os/arch`
   - **Node**: split on `-` as `os-arch`
   - **Rust**: split on `-` as `arch-vendor-os[-abi]`
   - **RubyGems**: split on `-` as `arch-os[-abi]`
   - **Python**: match against `manylinux_M_m_arch`, `musllinux_M_m_arch`, `macosx_M_m_arch`, `win_arch`, `win32`, or `linux_arch`
   - **Debian**: split on `-` as `arch-os-abi`
   - **LLVM**: split on `-` as `arch-vendor-os[-abi]`
   - **NuGet**: split on `-` as `os-arch` or `os-musl-arch`
   - **vcpkg**: split on `-` as `arch-os`
   - **Conan**: split on `/` or `-` as `os/arch` or `os-arch`
   - **Homebrew**: split on `_` as `arch_codename` (always macOS)
   - **Swift**: same as LLVM (`arch-vendor-os[-abi]`)
   - **Kotlin**: split camelCase as `osArch` (e.g., `linuxX64`)
   - **Maven**: split on `-` as `os-arch`
3. Resolve each component through the arch/OS reverse lookup indices
4. Return `ErrUnknownPlatform` if neither approach works

## Format rules

1. Look up the Platform in `platforms.json` by `arch/os/vendor/abi` key, trying default vendor and ABI if not specified
2. If found, return the preferred string for the target ecosystem (applying version overrides for Python)
3. If not found, compose from component mappings using ecosystem-specific rules:
   - Missing vendor defaults to `apple` for darwin/ios, `pc` for windows, `unknown` otherwise
   - Missing ABI defaults to `gnu` for linux, `msvc` for windows
4. Return `ErrNoMapping` if the ecosystem doesn't support the arch or OS

## Translate rules

`Translate(from, to, s)` parses from the source ecosystem, fills in default vendor and ABI, then formats for the target ecosystem. `Normalize(eco, s)` is `Translate(eco, eco, s)`.

## Information loss

Some translations are lossy. Go and Node cannot express ABI, so translating `x86_64-unknown-linux-musl` from Rust to Go gives `linux/amd64` with no way to recover the musl distinction. Python encodes OS and libc versions that other ecosystems ignore. Debian has no macOS or Windows support.
