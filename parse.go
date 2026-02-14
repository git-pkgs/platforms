package platforms

import (
	"fmt"
	"regexp"
	"strings"
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
	}
	return Platform{}, false
}

// go: os/arch
func decomposeGo(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
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
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
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
	parts := strings.SplitN(s, "-", 4)
	if len(parts) < 3 {
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
	if len(parts) == 4 {
		p.ABI = normalizeABI(parts[3])
	}
	return p, true
}

// rubygems: arch-os[-abi]  or cpu-os[-version]
func decomposeRubyGems(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", 3)
	if len(parts) < 2 {
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
	if len(parts) == 3 {
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
		abi := "gnu"
		if m[1] == "musl" {
			abi = "musl"
		}
		return Platform{
			Arch:        arch,
			OS:          "linux",
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
			OS:        "darwin",
			Vendor:    "apple",
			OSVersion: m[1] + "." + m[2],
		}, true
	}

	if s == "win32" {
		return Platform{Arch: "i686", OS: "windows", Vendor: "pc"}, true
	}

	if m := reWinPython.FindStringSubmatch(s); m != nil && m[1] != "" {
		arch := resolveArch(idx, Python, m[1])
		if arch == "" {
			return Platform{}, false
		}
		return Platform{Arch: arch, OS: "windows", Vendor: "pc"}, true
	}

	if m := reLinuxPython.FindStringSubmatch(s); m != nil {
		arch := resolveArch(idx, Python, m[1])
		if arch == "" {
			return Platform{}, false
		}
		return Platform{Arch: arch, OS: "linux"}, true
	}

	return Platform{}, false
}

// debian: arch-os-abi
func decomposeDebian(idx *indices, s string) (Platform, bool) {
	parts := strings.SplitN(s, "-", 3)
	if len(parts) != 3 {
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
	case s == "gnu" || s == "gnueabi":
		return "gnu"
	case s == "gnueabihf":
		return "eabihf"
	case s == "musl":
		return "musl"
	case s == "msvc":
		return "msvc"
	case strings.HasPrefix(s, "mingw"):
		return "mingw"
	case s == "eabi":
		return "eabi"
	case s == "eabihf":
		return "eabihf"
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

func validEcosystem(eco Ecosystem) bool {
	for _, e := range allEcosystems {
		if e == eco {
			return true
		}
	}
	return false
}
