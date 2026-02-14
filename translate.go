package platforms

// Translate parses a platform string from one ecosystem and formats it for another.
func Translate(from, to Ecosystem, s string) (string, error) {
	p, err := Parse(from, s)
	if err != nil {
		return "", err
	}

	// When translating to an ecosystem that needs vendor/ABI info but the
	// source ecosystem didn't provide it, fill in defaults.
	if p.Vendor == "" {
		p.Vendor = defaultVendor(p.OS)
	}
	if p.ABI == "" && p.OS == "linux" {
		p.ABI = "gnu"
	}

	return Format(to, p)
}

// Normalize parses and re-formats a platform string within the same ecosystem,
// returning the preferred form.
func Normalize(eco Ecosystem, s string) (string, error) {
	p, err := Parse(eco, s)
	if err != nil {
		return "", err
	}
	return Format(eco, p)
}
