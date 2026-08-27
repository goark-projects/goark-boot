package boot

import (
	"context"
	"sort"
	"strings"

	"goark.dev/goark"
	"goark.dev/goark/core/util"
	arkerrors "goark.dev/goark/errors"
)

// AutoConfiguration 表示 Boot 自动配置单元。
type AutoConfiguration interface {
	Name() string
	Order() int
	Configure(ctx context.Context, app *goark.ApplicationContext) error
}

// AutoConfigureFunc 是自动配置函数。
type AutoConfigureFunc func(ctx context.Context, app *goark.ApplicationContext) error

type autoConfiguration struct {
	name  string
	order int
	fn    AutoConfigureFunc
}

// AutoConfigurationOption 定制自动配置元数据。
type AutoConfigurationOption func(*autoConfiguration)

// NewAutoConfiguration 创建函数型自动配置。
func NewAutoConfiguration(name string, fn AutoConfigureFunc, options ...AutoConfigurationOption) AutoConfiguration {
	cfg := &autoConfiguration{name: strings.TrimSpace(name), fn: fn}
	for _, option := range options {
		if option != nil {
			option(cfg)
		}
	}
	return cfg
}

// WithAutoConfigurationOrder 设置自动配置顺序。
func WithAutoConfigurationOrder(order int) AutoConfigurationOption {
	return func(cfg *autoConfiguration) {
		cfg.order = order
	}
}

func (c *autoConfiguration) Name() string {
	return c.name
}

func (c *autoConfiguration) Order() int {
	return c.order
}

func (c *autoConfiguration) Configure(ctx context.Context, app *goark.ApplicationContext) error {
	if c == nil || c.fn == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "auto configuration is nil")
	}
	if c.name == "" {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "auto configuration name is empty")
	}
	return c.fn(ctx, app)
}

func configureApplication(ctx context.Context, app *goark.ApplicationContext, autoConfigurations []AutoConfiguration, configurations []goark.Configuration) error {
	for _, cfg := range sortedAutoConfigurations(autoConfigurations) {
		if util.IsNil(cfg) {
			continue
		}
		if err := ctx.Err(); err != nil {
			return arkerrors.Wrap(arkerrors.CodeLifecycle, err, "auto configuration canceled")
		}
		if err := cfg.Configure(ctx, app); err != nil {
			return arkerrors.Wrapf(arkerrors.CodeCreation, err, "auto configuration %q failed", cfg.Name())
		}
	}
	for _, configuration := range configurations {
		if util.IsNil(configuration) {
			continue
		}
		if err := app.RegisterConfiguration(configuration); err != nil {
			return err
		}
	}
	return nil
}

func sortedAutoConfigurations(configurations []AutoConfiguration) []AutoConfiguration {
	copied := append([]AutoConfiguration(nil), configurations...)
	sort.SliceStable(copied, func(i, j int) bool {
		if copied[i].Order() == copied[j].Order() {
			return copied[i].Name() < copied[j].Name()
		}
		return copied[i].Order() < copied[j].Order()
	})
	return copied
}
