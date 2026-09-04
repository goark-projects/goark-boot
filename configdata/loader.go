package configdata

import (
	"context"
	"strings"

	coreenv "goark.dev/goark/core/env"
	arkerrors "goark.dev/goark/errors"
)

// Loader 按 Boot 约定加载配置文件。
type Loader struct {
	options Options
	parser  Parser
}

// Result 表示配置加载结果。
type Result struct {
	Environment *coreenv.StandardEnvironment
	Sources     []LoadedSource
	Profiles    []string
}

// NewLoader 创建配置加载器。
func NewLoader(options ...Option) (*Loader, error) {
	config, err := newOptions(options...)
	if err != nil {
		return nil, err
	}
	return &Loader{
		options: config,
		parser:  NewViperParser(),
	}, nil
}

// Load 按默认规则加载配置文件。
func Load(ctx context.Context, options ...Option) (*Result, error) {
	loader, err := NewLoader(options...)
	if err != nil {
		return nil, err
	}
	return loader.Load(ctx)
}

// Load 加载基础配置和环境配置。
func (l *Loader) Load(ctx context.Context) (*Result, error) {
	if l == nil {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config loader is nil")
	}
	if ctx == nil {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "context is nil")
	}
	if l.parser == nil {
		return nil, arkerrors.New(arkerrors.CodeInvalidArgument, "config parser is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, arkerrors.Wrap(arkerrors.CodeLifecycle, err, "config loading canceled")
	}

	env, err := coreenv.NewEnvironment()
	if err != nil {
		return nil, err
	}
	if err := l.addSystemPropertiesSource(env); err != nil {
		return nil, err
	}
	if err := l.addCommandLineSource(env); err != nil {
		return nil, err
	}
	baseSources, err := l.loadBaseSources(ctx, env)
	if err != nil {
		return nil, err
	}
	profiles, err := l.resolveProfiles(env)
	if err != nil {
		return nil, err
	}
	if err := applyActiveProfiles(env, profiles); err != nil {
		return nil, err
	}
	profileSources, err := l.loadProfileSources(ctx, env, profiles)
	if err != nil {
		return nil, err
	}

	sources := append([]LoadedSource(nil), baseSources...)
	sources = append(sources, profileSources...)
	source, err := builtInDefaultPropertySource()
	if err != nil {
		return nil, err
	}
	if err := env.PropertySources().AddLast(source); err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		sources = append(sources, builtInLoadedSource())
	}
	return &Result{
		Environment: env,
		Sources:     sources,
		Profiles:    append([]string(nil), profiles...),
	}, nil
}

func (l *Loader) addSystemPropertiesSource(env *coreenv.StandardEnvironment) error {
	if len(l.options.SystemProperties) == 0 {
		return nil
	}
	source, err := coreenv.NewMapPropertySource(coreenv.SystemPropertiesPropertySourceName, propertyMap(l.options.SystemProperties))
	if err != nil {
		return err
	}
	return env.PropertySources().Replace(coreenv.SystemPropertiesPropertySourceName, source)
}

func (l *Loader) loadBaseSources(ctx context.Context, env *coreenv.StandardEnvironment) ([]LoadedSource, error) {
	candidates := discoverBaseFiles(l.options)
	return l.loadCandidates(ctx, env, candidates)
}

func (l *Loader) loadProfileSources(ctx context.Context, env *coreenv.StandardEnvironment, profiles []string) ([]LoadedSource, error) {
	candidates := discoverProfileFiles(l.options, profiles)
	return l.loadCandidates(ctx, env, candidates)
}

func (l *Loader) loadCandidates(ctx context.Context, env *coreenv.StandardEnvironment, candidates []Candidate) ([]LoadedSource, error) {
	loaded := make([]LoadedSource, 0, len(candidates))
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, arkerrors.Wrap(arkerrors.CodeLifecycle, err, "config loading canceled")
		}
		values, err := l.parser.ParseFile(candidate.Path, candidate.Format)
		if err != nil {
			return nil, err
		}
		source, err := coreenv.NewConfigPropertySource(candidate.SourceName(), propertyMap(values))
		if err != nil {
			return nil, err
		}
		if err := env.PropertySources().AddAfter(coreenv.SystemEnvironmentPropertySourceName, source); err != nil {
			return nil, err
		}
		loaded = append(loaded, candidate.LoadedSource())
	}
	return loaded, nil
}

func (l *Loader) resolveProfiles(env *coreenv.StandardEnvironment) ([]string, error) {
	if l.options.ProfilesExplicit {
		return expandProfiles(env, l.options.Profiles)
	}
	values := make([]string, 0, 1)
	for _, key := range []string{PropertyProfilesActive} {
		if value, ok := env.GetProperty(key); ok {
			values = append(values, value)
			break
		}
	}
	active, err := normalizeProfiles(values)
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		if value, ok := env.GetProperty(PropertyProfilesDefault); ok {
			active, err = normalizeProfiles([]string{value})
			if err != nil {
				return nil, err
			}
		}
	}
	if value, ok := env.GetProperty(PropertyProfilesInclude); ok {
		included, includeErr := normalizeProfiles([]string{value})
		if includeErr != nil {
			return nil, includeErr
		}
		active = appendUnique(active, included...)
	}
	return expandProfiles(env, active)
}

func (l *Loader) addCommandLineSource(env *coreenv.StandardEnvironment) error {
	if len(l.options.CommandLineProperties) == 0 {
		return nil
	}
	source, err := coreenv.NewMapPropertySource("commandLineArgs", propertyMap(l.options.CommandLineProperties))
	if err != nil {
		return err
	}
	return env.PropertySources().AddFirst(source)
}

func expandProfiles(env *coreenv.StandardEnvironment, profiles []string) ([]string, error) {
	expanded := make([]string, 0, len(profiles))
	visiting := make(map[string]bool)
	var add func(string) error
	add = func(profile string) error {
		if visiting[profile] {
			return arkerrors.Newf(arkerrors.CodeInvalidArgument, "circular profile group involving %q", profile)
		}
		if containsProfile(expanded, profile) {
			return nil
		}
		visiting[profile] = true
		expanded = append(expanded, profile)
		if value, ok := env.GetProperty(PropertyProfilesGroupPrefix + profile); ok {
			members, err := normalizeProfiles([]string{value})
			if err != nil {
				return err
			}
			for _, member := range members {
				if err := add(member); err != nil {
					return err
				}
			}
		}
		visiting[profile] = false
		return nil
	}
	for _, profile := range profiles {
		if err := add(strings.TrimSpace(profile)); err != nil {
			return nil, err
		}
	}
	return expanded, nil
}

func appendUnique(values []string, additions ...string) []string {
	for _, value := range additions {
		if !containsProfile(values, value) {
			values = append(values, value)
		}
	}
	return values
}

func containsProfile(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func applyActiveProfiles(env *coreenv.StandardEnvironment, profiles []string) error {
	for _, profile := range profiles {
		if err := env.AddActiveProfile(profile); err != nil {
			return err
		}
	}
	return nil
}

func propertyMap(values map[string]string) map[string]any {
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
