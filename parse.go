package platforms

import (
	"fmt"
	"regexp"
	"strings"
)

// Split limits for strings.SplitN in decompose functions.
const (
	splitTwo   = 2
	splitThree = 3
	splitFour  = 4
)

var (
	// manylinux_2_17_x86_64, musllinux_1_1_aarch64
	reManylinux = regexp.MustCompile(`^(many|musl)linux_(\d+)_(\d+)_(\w+)$`)
	// macosx_11_0_arm64
	reMacOSX = regexp.MustCompile(`^macosx_(\d+)_(\d+)_(\w+)$`)
	// win_amd64, win32, win_arm64
	reWinPython = regexp.MustCompile(`^win(?:_(\w+)|32)$`)
	// linux_x86_64, linux_aarch64, linux_armv7l
	reLinuxPython = regexp.MustCompile(`^linux_(\w+)$`)
)

// Parse converts an ecosystem-specific platform string into a canonical Platform.
func Parse(eco Ecosystem, s string) (Platform, error) {
	if !validEcosystem(eco) {
		return Platform{}, &ErrUnknownEcosystem{Ecosystem: string(eco)}
	}

	idx, err := loadData()
	if err != nil {
		return Platform{}, fmt.Errorf("loading platform data: %w", err)
	}

	// Check pre-computed platforms.json index first.
	if ecoIdx, ok := idx.platByString[string(eco)]; ok {
		if i, ok := ecoIdx[strings.ToLower(s)]; ok {
			p := idx.platFile.Platforms[i]
			plat := Platform{
				Arch:   p.Arch,
				OS:     p.OS,
				Vendor: p.Vendor,
				ABI:    p.ABI,
			}
			// Extract version info from Python strings.
			if eco == Python {
				parsePythonVersions(s, &plat)
			}
			return plat, nil
		}
	}

	// Decompose using ecosystem-specific rules.
	plat, ok := decompose(idx, eco, s)
	if ok {
		return plat, nil
	}

	return Platform{}, &ErrUnknownPlatform{Ecosystem: eco, Input: s}
}

func decompose(idx *indices, eco Ecosystem, s string) (Platform, bool) {
	switch eco {
	case Go:
		return decomposeGo(idx, s)
	case Node:
		return decomposeNode(idx, s)
	case Rust:
		return decomposeRustLLVM(idx, eco, s)
	case LLVM:
		return decomposeRustLLVM(idx, eco, s)
	case RubyGems:
		return decomposeRubyGems(idx, s)
	case Python:
		return decomposePython(idx, s)
	case Debian:
		return decomposeDebian(idx, s)
	case NuGet:
		return decomposeNuGet(idx, s)
	case Vcpkg:
		return decomposeVcpkg(idx, s)
	case Conan:
		return decomposeConan(idx, s)
	case Homebrew:
		return decomposeHomebrew(idx, s)
	case Swift:
		return decomposeRustLLVM(idx, Swift, s)
	case Kotlin:
		return decomposeKotlin(idx, s)
	case Maven:
		return decomposeMaven(idx, s)
	}
	return Platform{}, false
}

// go: os/arch
func decomposeGo(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "/", splitTwo)
	if len(parts) != splitTwo {
		return Platform{}, false
	}
	osName := resolveOS(idx, Go, parts[0])
	arch := resolveArch(idx, Go, parts[1])
	if osName == "" || arch == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

// node: os-arch
func decomposeNode(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitTwo)
	if len(parts) != splitTwo {
		return Platform{}, false
	}
	osName := resolveOS(idx, Node, parts[0])
	arch := resolveArch(idx, Node, parts[1])
	if osName == "" || arch == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

// rust/llvm: arch-vendor-os[-abi]
func decomposeRustLLVM(idx *indices, eco Ecosystem, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitFour)
	if len(parts) < splitThree {
		return Platform{}, false
	}
	arch := resolveArch(idx, eco, parts[0])
	if arch == "" {
		return Platform{}, false
	}
	vendor := parts[1]
	osName := resolveOS(idx, eco, parts[2])
	if osName == "" {
		return Platform{}, false
	}
	p := Platform{Arch: arch, OS: osName, Vendor: vendor}
	if len(parts) == splitFour {
		p.ABI = normalizeABI(parts[3])
	}
	return p, true
}

// rubygems: arch-os[-abi]  or cpu-os[-version]
func decomposeRubyGems(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitThree)
	if len(parts) < splitTwo {
		return Platform{}, false
	}
	arch := resolveArch(idx, RubyGems, parts[0])
	if arch == "" {
		return Platform{}, false
	}
	osName := resolveOS(idx, RubyGems, parts[1])
	if osName == "" {
		return Platform{}, false
	}
	p := Platform{Arch: arch, OS: osName}
	if len(parts) == splitThree {
		p.ABI = normalizeABI(parts[2])
	}
	return p, true
}

// python: manylinux_M_m_arch, musllinux_M_m_arch, macosx_M_m_arch, win_arch, win32, linux_arch
func decomposePython(idx *indices, s string) (Platform, bool) {
	if m := reManylinux.FindStringSubmatch(s); m != nil {
		arch := resolveArch(idx, Python, m[4])
		if arch == "" {
			return Platform{}, false
		}
		abi := abiGNU
		if m[1] == abiMusl {
			abi = abiMusl
		}
		return Platform{
			Arch:        arch,
			OS:          osLinux,
			ABI:         abi,
			LibCVersion: m[2] + "." + m[3],
		}, true
	}

	if m := reMacOSX.FindStringSubmatch(s); m != nil {
		arch := resolveArch(idx, Python, m[3])
		if arch == "" {
			return Platform{}, false
		}
		return Platform{
			Arch:      arch,
			OS:        osDarwin,
			Vendor:    "apple",
			OSVersion: m[1] + "." + m[2],
		}, true
	}

	if s == "win32" {
		return Platform{Arch: "i686", OS: osWindows, Vendor: "pc"}, true
	}

	if m := reWinPython.FindStringSubmatch(s); m != nil && m[1] != "" {
		arch := resolveArch(idx, Python, m[1])
		if arch == "" {
			return Platform{}, false
		}
		return Platform{Arch: arch, OS: osWindows, Vendor: "pc"}, true
	}

	if m := reLinuxPython.FindStringSubmatch(s); m != nil {
		arch := resolveArch(idx, Python, m[1])
		if arch == "" {
			return Platform{}, false
		}
		return Platform{Arch: arch, OS: osLinux}, true
	}

	return Platform{}, false
}

// debian: arch-os-abi
func decomposeDebian(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitThree)
	if len(parts) != splitThree {
		return Platform{}, false
	}
	arch := resolveArch(idx, Debian, parts[0])
	if arch == "" {
		return Platform{}, false
	}
	osName := resolveOS(idx, Debian, parts[1])
	if osName == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName, ABI: normalizeABI(parts[2])}, true
}

func normalizeABI(s string) string {
	s = strings.ToLower(s)
	switch {
	case s == abiGNU || s == abiGNUEABI:
		return abiGNU
	case s == abiGNUEABIHF:
		return abiEABIHF
	case s == abiMusl:
		return abiMusl
	case s == abiMSVC:
		return abiMSVC
	case strings.HasPrefix(s, "mingw"):
		return "mingw"
	case s == abiEABI:
		return abiEABI
	case s == abiEABIHF:
		return abiEABIHF
	}
	return s
}

func parsePythonVersions(s string, p *Platform) {
	if m := reManylinux.FindStringSubmatch(s); m != nil {
		p.LibCVersion = m[2] + "." + m[3]
	} else if m := reMacOSX.FindStringSubmatch(s); m != nil {
		p.OSVersion = m[1] + "." + m[2]
	}
}

// nuget: os-arch or os-musl-arch (e.g., linux-x64, linux-musl-x64, win-arm64, osx-x64)
func decomposeNuGet(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitThree)
	if len(parts) == splitThree && strings.ToLower(parts[1]) == abiMusl {
		osName := resolveOS(idx, NuGet, parts[0])
		arch := resolveArch(idx, NuGet, parts[2])
		if osName == "" || arch == "" {
			return Platform{}, false
		}
		return Platform{Arch: arch, OS: osName, ABI: abiMusl}, true
	}
	if len(parts) < splitTwo {
		return Platform{}, false
	}
	osName := resolveOS(idx, NuGet, parts[0])
	arch := resolveArch(idx, NuGet, parts[1])
	if osName == "" || arch == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

// vcpkg: arch-os (e.g., x64-linux, arm64-osx, x64-windows)
func decomposeVcpkg(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitTwo)
	if len(parts) != splitTwo {
		return Platform{}, false
	}
	arch := resolveArch(idx, Vcpkg, parts[0])
	osName := resolveOS(idx, Vcpkg, parts[1])
	if arch == "" || osName == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

// conan: structured settings, but we handle "os-arch" style strings
// Conan uses settings like os=Linux, arch=armv8 but we parse "os/arch" pairs
func decomposeConan(idx *indices, s string) (Platform, bool) {
	// Try os/arch with slash separator
	parts := strings.SplitN(s, "/", splitTwo)
	if len(parts) == splitTwo {
		osName := resolveOS(idx, Conan, parts[0])
		arch := resolveArch(idx, Conan, parts[1])
		if osName != "" && arch != "" {
			return Platform{Arch: arch, OS: osName}, true
		}
	}
	// Try os-arch with dash separator
	parts = strings.SplitN(s, "-", splitTwo)
	if len(parts) == splitTwo {
		osName := resolveOS(idx, Conan, parts[0])
		arch := resolveArch(idx, Conan, parts[1])
		if osName != "" && arch != "" {
			return Platform{Arch: arch, OS: osName}, true
		}
	}
	return Platform{}, false
}

// homebrew: arch_codename (e.g., arm64_sonoma, arm64_ventura)
// We can only extract the arch since the codename maps to a macOS version, not a canonical OS
func decomposeHomebrew(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "_", splitTwo)
	if len(parts) != splitTwo {
		return Platform{}, false
	}
	arch := resolveArch(idx, Homebrew, parts[0])
	if arch == "" {
		return Platform{}, false
	}
	// Homebrew is macOS-only
	return Platform{Arch: arch, OS: osDarwin, Vendor: "apple"}, true
}

// kotlin: camelCase DSL names (e.g., linuxX64, macosArm64, mingwX64, iosArm64, androidNativeArm64)
var reKotlin = regexp.MustCompile(`^([a-z]+)([A-Z]\w+)$`)

func decomposeKotlin(idx *indices, s string) (Platform, bool) {
	m := reKotlin.FindStringSubmatch(s)
	if m == nil {
		return Platform{}, false
	}
	osName := resolveOS(idx, Kotlin, m[1])
	arch := resolveArch(idx, Kotlin, m[2])
	if osName == "" || arch == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

// maven: os-arch (e.g., linux-x86_64, osx-aarch_64, windows-x86_64)
func decomposeMaven(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", splitTwo)
	if len(parts) != splitTwo {
		return Platform{}, false
	}
	osName := resolveOS(idx, Maven, parts[0])
	arch := resolveArch(idx, Maven, parts[1])
	if osName == "" || arch == "" {
		return Platform{}, false
	}
	return Platform{Arch: arch, OS: osName}, true
}

func validEcosystem(eco Ecosystem) bool {
	for _, e := range allEcosystems {
		if e == eco {
			return true
		}
	}
	return false
}
