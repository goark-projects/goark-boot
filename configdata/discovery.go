package configdata

import (
	"os"
	"path/filepath"
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
}

func discoverBaseFiles(options Options) []Candidate {
	candidates := make([]Candidate, 0, len(options.Locations))
	for _, location := range options.Locations {
		if candidate, ok := firstExistingCandidate(options.BaseName, "", location, options.Formats); ok {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func discoverProfileFiles(options Options, profiles []string) []Candidate {
	candidates := make([]Candidate, 0, len(options.Locations)*len(profiles))
	for _, profile := range profiles {
		for _, location := range options.Locations {
			name := options.BaseName + "-" + profile
			if candidate, ok := firstExistingCandidate(name, profile, location, options.Formats); ok {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

func firstExistingCandidate(name string, profile string, location string, formats []Format) (Candidate, bool) {
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
