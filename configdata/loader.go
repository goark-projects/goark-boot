package configdata

import (
	"context"

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
	if len(sources) == 0 {
		source, err := builtInDefaultPropertySource()
		if err != nil {
			return nil, err
		}
		if err := env.PropertySources().AddLast(source); err != nil {
			return nil, err
		}
		sources = append(sources, builtInLoadedSource())
	}
	return &Result{
		Environment: env,
		Sources:     sources,
		Profiles:    append([]string(nil), profiles...),
	}, nil
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
		if err := env.PropertySources().AddFirst(source); err != nil {
			return nil, err
		}
		loaded = append(loaded, candidate.LoadedSource())
	}
	return loaded, nil
}

func (l *Loader) resolveProfiles(env *coreenv.StandardEnvironment) ([]string, error) {
	if l.options.ProfilesExplicit {
		return append([]string(nil), l.options.Profiles...), nil
	}
	values := make([]string, 0, 1)
	for _, key := range []string{profileKeyGoark, profileKeySpring, profileKeyShort} {
		if value, ok := env.GetProperty(key); ok {
			values = append(values, value)
			break
		}
	}
	return normalizeProfiles(values)
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
