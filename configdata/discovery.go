package configdata

import (
	"os"
	"path/filepath"
	"strings"
)

// Candidate 表示一个待加载配置文件。
type Candidate struct {
	Path     string
	Format   Format
	Profile  string
	Location string
}

// LoadedSource 表示一个已加载配置源。
type LoadedSource struct {
	Name     string
	Path     string
	Format   Format
	Profile  string
	Location string
	BuiltIn  bool
}

func discoverBaseFiles(options Options) []Candidate {
	locations := append(append([]string(nil), options.Locations...), options.AdditionalLocations...)
	candidates := make([]Candidate, 0, len(locations))
	for _, location := range locations {
		if candidate, ok := firstExistingNamedCandidate(options.baseNames(), "", location, options.Formats); ok {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func discoverProfileFiles(options Options, profiles []string) []Candidate {
	locations := append(append([]string(nil), options.Locations...), options.AdditionalLocations...)
	candidates := make([]Candidate, 0, len(locations)*len(profiles))
	for _, profile := range profiles {
		for _, location := range locations {
			names := make([]string, 0, 2)
			for _, baseName := range options.baseNames() {
				names = append(names, baseName+"-"+profile)
			}
			if candidate, ok := firstExistingNamedCandidate(names, profile, location, options.Formats); ok {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

func firstExistingNamedCandidate(names []string, profile string, location string, formats []Format) (Candidate, bool) {
	for _, name := range names {
		if candidate, ok := firstExistingCandidate(name, profile, location, formats); ok {
			return candidate, true
		}
	}
	return Candidate{}, false
}

func firstExistingCandidate(name string, profile string, location string, formats []Format) (Candidate, bool) {
	if format, ok := explicitFileFormat(location, formats); ok {
		path := location
		if profile != "" {
			path = profileFilePath(location, profile)
		}
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return Candidate{
				Path:     filepath.Clean(path),
				Format:   format,
				Profile:  profile,
				Location: filepath.Clean(filepath.Dir(path)),
			}, true
		}
		return Candidate{}, false
	}
	for _, format := range formats {
		path := filepath.Join(location, name+"."+string(format))
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return Candidate{
				Path:     filepath.Clean(path),
				Format:   format,
				Profile:  profile,
				Location: filepath.Clean(location),
			}, true
		}
	}
	return Candidate{}, false
}

func explicitFileFormat(location string, formats []Format) (Format, bool) {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(location), "."))
	if extension == "" {
		return "", false
	}
	for _, format := range formats {
		if string(format) == extension {
			return format, true
		}
	}
	return "", false
}

func profileFilePath(path string, profile string) string {
	extension := filepath.Ext(path)
	return strings.TrimSuffix(path, extension) + "-" + profile + extension
}

func (c Candidate) SourceName() string {
	return "config:" + c.Path
}

func (c Candidate) LoadedSource() LoadedSource {
	return LoadedSource{
		Name:     c.SourceName(),
		Path:     c.Path,
		Format:   c.Format,
		Profile:  c.Profile,
		Location: c.Location,
	}
}
