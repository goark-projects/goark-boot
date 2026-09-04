package configdata

import (
	"os"
	"path/filepath"
	"strings"

	arkerrors "goark.dev/goark/errors"
)

const (
	defaultBaseName = "app"
	defaultResource = "resource"

	profileKeyGoark = "goark.profiles.active"
)

const (
	// PropertyConfigLocation 设置配置文件或配置目录位置。
	PropertyConfigLocation = "goark.config.location"
	// PropertyConfigAdditionalLocation 增加配置文件或配置目录位置。
	PropertyConfigAdditionalLocation = "goark.config.additional-location"
	// PropertyConfigName 设置配置文件基础名称。
	PropertyConfigName = "goark.config.name"
	// PropertyProfilesActive 设置激活 Profile。
	PropertyProfilesActive = profileKeyGoark
	// PropertyProfilesInclude 设置附加激活的 Profile。
	PropertyProfilesInclude = "goark.profiles.include"
	// PropertyProfilesDefault 设置没有显式激活项时使用的 Profile。
	PropertyProfilesDefault = "goark.profiles.default"
	// PropertyProfilesGroupPrefix 是 Profile 组属性前缀。
	PropertyProfilesGroupPrefix = "goark.profiles.group."
	// PropertyApplicationName 设置应用名称。
	PropertyApplicationName = "goark.application.name"
)

const (
	// EnvConfigLocation 设置配置文件或配置目录位置。
	EnvConfigLocation = "GOARK_CONFIG_LOCATION"
	// EnvConfigAdditionalLocation 增加配置文件或配置目录位置。
	EnvConfigAdditionalLocation = "GOARK_CONFIG_ADDITIONAL_LOCATION"
	// EnvConfigName 设置配置文件基础名称。
	EnvConfigName = "GOARK_CONFIG_NAME"
	// EnvProfilesActive 设置激活 Profile。
	EnvProfilesActive = "GOARK_PROFILES_ACTIVE"
)

// Format 表示配置文件格式。
type Format string

const (
	FormatYAML       Format = "yml"
	FormatYAMLFull   Format = "yaml"
	FormatTOML       Format = "toml"
	FormatProperties Format = "properties"
)

// Options 描述配置文件加载规则。
type Options struct {
	BaseName              string
	BaseNameExplicit      bool
	Profiles              []string
	ProfilesExplicit      bool
	Locations             []string
	AdditionalLocations   []string
	Formats               []Format
	CommandLineProperties map[string]string
	SystemProperties      map[string]string
}

// Option 调整配置加载规则。
type Option func(*Options) error

// WithBaseName 设置配置文件基础名称。
func WithBaseName(name string) Option {
	return func(options *Options) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return arkerrors.New(arkerrors.CodeInvalidArgument, "config base name is empty")
		}
		if strings.ContainsAny(name, `/\`) {
			return arkerrors.Newf(arkerrors.CodeInvalidArgument, "config base name %q must not contain path separators", name)
		}
		options.BaseName = name
		options.BaseNameExplicit = true
		return nil
	}
}

// WithProfiles 设置激活环境，后声明的环境优先级更高。
func WithProfiles(profiles ...string) Option {
	return func(options *Options) error {
		normalized, err := normalizeProfiles(profiles)
		if err != nil {
			return err
		}
		options.Profiles = normalized
		options.ProfilesExplicit = true
		return nil
	}
}

// WithLocations 设置配置目录，越靠后的目录优先级越高。
func WithLocations(locations ...string) Option {
	return func(options *Options) error {
		normalized, err := normalizeLocations(locations)
		if err != nil {
			return err
		}
		options.Locations = normalized
		return nil
	}
}

// WithArgs 应用 Goark 命令行配置参数。
func WithArgs(args ...string) Option {
	copied := append([]string(nil), args...)
	return func(options *Options) error {
		return applyArguments(options, copied)
	}
}

// WithFormats 设置同名配置文件的格式优先级。
func WithFormats(formats ...Format) Option {
	return func(options *Options) error {
		normalized, err := normalizeFormats(formats)
		if err != nil {
			return err
		}
		options.Formats = normalized
		return nil
	}
}

func newOptions(options ...Option) (Options, error) {
	defaultLocations, err := defaultLocations()
	if err != nil {
		return Options{}, err
	}
	config := Options{
		BaseName:              defaultBaseName,
		Locations:             defaultLocations,
		Formats:               []Format{FormatYAML, FormatYAMLFull, FormatProperties, FormatTOML},
		CommandLineProperties: make(map[string]string),
		SystemProperties:      make(map[string]string),
	}
	if err := applyEnvironment(&config); err != nil {
		return Options{}, err
	}
	if err := applyArguments(&config, os.Args[1:]); err != nil {
		return Options{}, err
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&config); err != nil {
			return Options{}, err
		}
	}
	return config, nil
}

func (o Options) baseNames() []string {
	return []string{o.BaseName}
}

func (o Options) profileBaseNames() []string {
	return []string{o.BaseName}
}

func (o Options) formatsForBaseName(name string) []Format {
	allowed := map[Format]struct{}{FormatYAML: {}, FormatProperties: {}}
	if name == defaultBaseName || o.BaseNameExplicit {
		allowed[FormatTOML] = struct{}{}
	}
	formats := make([]Format, 0, len(allowed))
	for _, format := range o.Formats {
		if _, ok := allowed[format]; ok {
			formats = append(formats, format)
		}
	}
	return formats
}

func defaultLocations() ([]string, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, arkerrors.Wrap(arkerrors.CodeInvalidArgument, err, "failed to resolve executable path")
	}
	executableDir := filepath.Dir(executable)
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, arkerrors.Wrap(arkerrors.CodeInvalidArgument, err, "failed to resolve working directory")
	}
	return defaultLocationsFor(executableDir, workingDir)
}

func defaultLocationsFor(executableDir string, workingDir string) ([]string, error) {
	locations := []string{
		filepath.Join(executableDir, defaultResource),
	}
	if strings.TrimSpace(workingDir) != "" {
		locations = append(locations, filepath.Join(workingDir, defaultResource))
	}
	return normalizeLocations(locations)
}

func normalizeLocations(locations []string) ([]string, error) {
	if len(locations) == 0 {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config locations are empty")
	}
	rawLocations := splitLocations(locations)
	seen := make(map[string]struct{}, len(rawLocations))
	normalized := make([]string, 0, len(locations))
	for _, location := range rawLocations {
		location = strings.TrimSpace(location)
		if location == "" {
			return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config location is empty")
		}
		location = strings.TrimPrefix(location, "file:")
		absolute, err := filepath.Abs(location)
		if err != nil {
			return nil, arkerrors.Wrapf(arkerrors.CodeInvalidArgument, err, "failed to resolve config location %q", location)
		}
		absolute = filepath.Clean(absolute)
		if _, exists := seen[absolute]; exists {
			continue
		}
		seen[absolute] = struct{}{}
		normalized = append(normalized, absolute)
	}
	if len(normalized) == 0 {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config locations are empty")
	}
	return normalized, nil
}

func splitLocations(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			if item = strings.TrimSpace(item); item != "" {
				out = append(out, item)
			}
		}
	}
	return out
}

func normalizeProfiles(profiles []string) ([]string, error) {
	seen := make(map[string]struct{}, len(profiles))
	normalized := make([]string, 0, len(profiles))
	for _, profile := range splitProfiles(profiles) {
		if strings.ContainsAny(profile, `/\`) {
			return nil, arkerrors.Newf(arkerrors.CodeInvalidArgument, "profile %q must not contain path separators", profile)
		}
		if _, exists := seen[profile]; exists {
			continue
		}
		seen[profile] = struct{}{}
		normalized = append(normalized, profile)
	}
	return normalized, nil
}

func splitProfiles(values []string) []string {
	profiles := make([]string, 0, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				profiles = append(profiles, item)
			}
		}
	}
	return profiles
}

func normalizeFormats(formats []Format) ([]Format, error) {
	if len(formats) == 0 {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config formats are empty")
	}
	seen := make(map[Format]struct{}, len(formats))
	normalized := make([]Format, 0, len(formats))
	for _, format := range formats {
		format = Format(strings.ToLower(strings.TrimPrefix(strings.TrimSpace(string(format)), ".")))
		switch format {
		case FormatYAML, FormatYAMLFull, FormatTOML, FormatProperties:
		default:
			return nil, arkerrors.Newf(arkerrors.CodeInvalidArgument, "unsupported config format %q", format)
		}
		if _, exists := seen[format]; exists {
			continue
		}
		seen[format] = struct{}{}
		normalized = append(normalized, format)
	}
	return normalized, nil
}
