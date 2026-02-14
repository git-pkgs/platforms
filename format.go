package platforms

import (
	"fmt"
	"strings"
)

// Format converts a canonical Platform into an ecosystem-specific string.
func Format(eco Ecosystem, p Platform) (string, error) {
	if !validEcosystem(eco) {
		return "", &ErrUnknownEcosystem{Ecosystem: string(eco)}
	}

	idx, err := loadData()
	if err != nil {
		return "", fmt.Errorf("loading platform data: %w", err)
	}

	// Try pre-computed platforms.json index with several key strategies.
	if s, ok := formatFromPrecomputed(idx, eco, p); ok {
		return s, nil
	}

	// Compose from component mappings.
	return compose(idx, eco, p)
}

func formatFromPrecomputed(idx *indices, eco Ecosystem, p Platform) (string, bool) {
	// Build a list of keys to try, from most specific to least.
	keys := []string{platKey(p.Arch, p.OS, p.Vendor, p.ABI)}

	// Try with default vendor filled in.
	dv := defaultVendor(p.OS)
	if p.Vendor == "" {
		keys = append(keys, platKey(p.Arch, p.OS, dv, p.ABI))
	}

	// Try with default ABI filled in.
	da := defaultABI(p.OS)
	if p.ABI == "" && da != "" {
		keys = append(keys, platKey(p.Arch, p.OS, p.Vendor, da))
		if p.Vendor == "" {
			keys = append(keys, platKey(p.Arch, p.OS, dv, da))
		}
	}

	for _, key := range keys {
		if i, ok := idx.platByKey[key]; ok {
			entry := idx.platFile.Platforms[i]
			if raw, ok := entry.Strings[string(eco)]; ok {
				pref, _ := resolveMapping(raw)
				if pref != "" {
					return applyVersions(eco, pref, p), true
				}
			}
		}
	}
	return "", false
}

func compose(idx *indices, eco Ecosystem, p Platform) (string, error) {
	arch := lookupArch(idx, eco, p.Arch)
	osName := lookupOS(idx, eco, p.OS)

	if arch == "" || osName == "" {
		return "", &ErrNoMapping{Ecosystem: eco, Platform: p}
	}

	switch eco {
	case Go:
		return osName + "/" + arch, nil
	case Node:
		return osName + "-" + arch, nil
	case Rust:
		vendor := p.Vendor
		if vendor == "" {
			vendor = defaultVendor(p.OS)
		}
		abi := p.ABI
		if abi == "" && p.OS == "linux" {
			abi = "gnu"
		}
		if abi != "" {
			return arch + "-" + vendor + "-" + osName + "-" + rustABI(abi), nil
		}
		return arch + "-" + vendor + "-" + osName, nil
	case RubyGems:
		if p.ABI != "" && p.ABI != "gnu" {
			return arch + "-" + osName + "-" + p.ABI, nil
		}
		return arch + "-" + osName, nil
	case Python:
		return composePython(arch, osName, p), nil
	case Debian:
		abi := p.ABI
		if abi == "" && p.OS == "linux" {
			abi = "gnu"
		}
		if abi == "" {
			return "", &ErrNoMapping{Ecosystem: eco, Platform: p}
		}
		return arch + "-" + osName + "-" + debianABI(abi), nil
	case LLVM:
		vendor := p.Vendor
		if vendor == "" {
			vendor = defaultVendor(p.OS)
		}
		abi := p.ABI
		if abi == "" && p.OS == "linux" {
			abi = "gnu"
		}
		if abi != "" {
			return arch + "-" + vendor + "-" + osName + "-" + abi, nil
		}
		return arch + "-" + vendor + "-" + osName, nil
	case NuGet:
		if p.ABI == "musl" {
			return osName + "-musl-" + arch, nil
		}
		return osName + "-" + arch, nil
	case Vcpkg:
		return arch + "-" + osName, nil
	case Conan:
		return osName + "/" + arch, nil
	case Homebrew:
		if p.OS != "darwin" {
			return "", &ErrNoMapping{Ecosystem: eco, Platform: p}
		}
		return arch + "_darwin", nil
	case Swift:
		vendor := p.Vendor
		if vendor == "" {
			vendor = defaultVendor(p.OS)
		}
		abi := p.ABI
		if abi == "" && p.OS == "linux" {
			abi = "gnu"
		}
		if abi != "" {
			return arch + "-" + vendor + "-" + osName + "-" + abi, nil
		}
		return arch + "-" + vendor + "-" + osName, nil
	case Kotlin:
		return osName + arch, nil
	case Maven:
		return osName + "-" + arch, nil
	}
	return "", &ErrNoMapping{Ecosystem: eco, Platform: p}
}

func defaultVendor(os string) string {
	switch os {
	case "darwin", "ios":
		return "apple"
	case "windows":
		return "pc"
	default:
		return "unknown"
	}
}

func defaultABI(os string) string {
	switch os {
	case "linux":
		return "gnu"
	case "windows":
		return "msvc"
	default:
		return ""
	}
}

func rustABI(abi string) string {
	switch abi {
	case "eabihf":
		return "gnueabihf"
	case "eabi":
		return "gnueabi"
	default:
		return abi
	}
}

func debianABI(abi string) string {
	switch abi {
	case "eabihf":
		return "gnueabihf"
	case "eabi":
		return "gnueabi"
	case "gnu":
		return "gnu"
	default:
		return abi
	}
}

func composePython(arch, osName string, p Platform) string {
	switch p.OS {
	case "darwin":
		ver := p.OSVersion
		if ver == "" {
			if p.Arch == "aarch64" {
				ver = "11.0"
			} else {
				ver = "10.9"
			}
		}
		verParts := underscoreVersion(ver)
		return "macosx_" + verParts + "_" + arch
	case "linux":
		if p.ABI == "musl" {
			ver := p.LibCVersion
			if ver == "" {
				ver = "1.1"
			}
			verParts := underscoreVersion(ver)
			return "musllinux_" + verParts + "_" + arch
		}
		ver := p.LibCVersion
		if ver == "" {
			ver = "2.17"
		}
		verParts := underscoreVersion(ver)
		return "manylinux_" + verParts + "_" + arch
	case "windows":
		if p.Arch == "i686" {
			return "win32"
		}
		return "win_" + arch
	}
	return osName + "_" + arch
}

func underscoreVersion(v string) string {
	return strings.ReplaceAll(v, ".", "_")
}

func applyVersions(eco Ecosystem, base string, p Platform) string {
	if eco != Python {
		return base
	}

	// For Python manylinux/musllinux, replace the version in the string.
	if p.LibCVersion != "" && reManylinux.MatchString(base) {
		m := reManylinux.FindStringSubmatch(base)
		prefix := m[1] + "linux"
		verParts := underscoreVersion(p.LibCVersion)
		return prefix + "_" + verParts + "_" + m[4]
	}
	if p.OSVersion != "" && reMacOSX.MatchString(base) {
		m := reMacOSX.FindStringSubmatch(base)
		verParts := underscoreVersion(p.OSVersion)
		return "macosx_" + verParts + "_" + m[3]
	}
	return base
}
