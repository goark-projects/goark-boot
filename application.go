package boot

import (
	"context"
	"sync"

	"goark.dev/boot/configdata"
	"goark.dev/goark"
	arkerrors "goark.dev/goark/errors"
)

// Application 表示一个 Goark Boot 应用实例。
type Application struct {
	mu                 sync.Mutex
	context            *goark.ApplicationContext
	contextOptions     []goark.Option
	configDataOptions  []configdata.Option
	configurations     []goark.Configuration
	autoConfigurations []AutoConfiguration
	configData         *configdata.Result
	started            bool
}

// New 创建 Boot 应用。
func New(options ...Option) (*Application, error) {
	app := &Application{}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(app); err != nil {
			return nil, err
		}
	}
	return app, nil
}

// MustNew 创建 Boot 应用，失败时 panic。
func MustNew(options ...Option) *Application {
	app, err := New(options...)
	if err != nil {
		panic(err)
	}
	return app
}

// Run 创建并启动 Boot 应用。
func Run(ctx context.Context, options ...Option) (*Application, error) {
	app, err := New(options...)
	if err != nil {
		return nil, err
	}
	if err := app.Run(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

// Run 启动当前应用实例。
func (a *Application) Run(ctx context.Context) error {
	if a == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "boot application is nil")
	}
	if ctx == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "context is nil")
	}

	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	configData, err := configdata.Load(ctx, a.configDataOptions...)
	if err != nil {
		return err
	}
	contextOptions := make([]goark.Option, 0, len(a.contextOptions)+1)
	contextOptions = append(contextOptions, goark.WithEnvironment(configData.Environment))
	contextOptions = append(contextOptions, a.contextOptions...)
	appContext, err := goark.New(contextOptions...)
	if err != nil {
		return err
	}
	if err := configureApplication(ctx, appContext, a.autoConfigurations, a.configurations); err != nil {
		return err
	}
	if err := appContext.Start(ctx); err != nil {
		return err
	}

	a.mu.Lock()
	a.context = appContext
	a.configData = configData
	a.started = true
	a.mu.Unlock()
	return nil
}

// Context 返回底层 Goark 应用上下文。
func (a *Application) Context() (*goark.ApplicationContext, bool) {
	if a == nil {
		return nil, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.context, a.context != nil
}

// ConfigData 返回配置加载结果。
func (a *Application) ConfigData() (*configdata.Result, bool) {
	if a == nil {
		return nil, false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.configData, a.configData != nil
}

// Close 关闭底层 Goark 应用上下文。
func (a *Application) Close(ctx context.Context) error {
	if a == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "boot application is nil")
	}
	if ctx == nil {
		return arkerrors.New(arkerrors.CodeInvalidArgument, "context is nil")
	}
	a.mu.Lock()
	appContext := a.context
	a.started = false
	a.mu.Unlock()
	if appContext == nil {
		return nil
	}
	return appContext.Close(ctx)
}
