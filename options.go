package boot

import (
	"goark.dev/boot/configdata"
	"goark.dev/goark"
)

// Option 定制 Boot 应用。
type Option func(*Application) error

// WithContextOptions 添加底层 Goark ApplicationContext 选项。
func WithContextOptions(options ...goark.Option) Option {
	return func(app *Application) error {
		app.contextOptions = append(app.contextOptions, options...)
		return nil
	}
}

// WithConfigDataOptions 添加配置数据加载选项。
func WithConfigDataOptions(options ...configdata.Option) Option {
	return func(app *Application) error {
		app.configDataOptions = append(app.configDataOptions, options...)
		return nil
	}
}

// WithConfiguration 注册用户配置单元。
func WithConfiguration(configurations ...goark.Configuration) Option {
	return func(app *Application) error {
		app.configurations = append(app.configurations, configurations...)
		return nil
	}
}

// WithAutoConfiguration 注册自动配置单元。
func WithAutoConfiguration(configurations ...AutoConfiguration) Option {
	return func(app *Application) error {
		app.autoConfigurations = append(app.autoConfigurations, configurations...)
		return nil
	}
}
