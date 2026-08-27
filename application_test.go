package boot_test

import (
	"context"
	"reflect"
	"testing"

	"goark.dev/boot"
	"goark.dev/goark"
	"goark.dev/goark/container"
)

func TestRunAppliesAutoConfigurationsInOrder(t *testing.T) {
	calls := make([]string, 0, 2)
	app, err := boot.Run(t.Context(),
		boot.WithAutoConfiguration(
			boot.NewAutoConfiguration("second", func(context.Context, *goark.ApplicationContext) error {
				calls = append(calls, "second")
				return nil
			}, boot.WithAutoConfigurationOrder(20)),
			boot.NewAutoConfiguration("first", func(context.Context, *goark.ApplicationContext) error {
				calls = append(calls, "first")
				return nil
			}, boot.WithAutoConfigurationOrder(10)),
		),
	)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	defer func() {
		if err := app.Close(t.Context()); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}()
	if !reflect.DeepEqual(calls, []string{"first", "second"}) {
		t.Fatalf("calls = %v", calls)
	}
}

func TestRunRegistersUserConfiguration(t *testing.T) {
	app, err := boot.Run(t.Context(), boot.WithConfiguration(testConfiguration{}))
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	defer func() {
		if err := app.Close(t.Context()); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}()

	ctx, ok := app.Context()
	if !ok {
		t.Fatal("application context not available")
	}
	value := goark.MustGet[string](t.Context(), ctx, "message")
	if value != "ok" {
		t.Fatalf("message = %q, want ok", value)
	}
}

type testConfiguration struct{}

func (testConfiguration) Name() string { return "test" }

func (testConfiguration) Order() int { return 0 }

func (testConfiguration) Register(_ context.Context, registry *container.Registry) error {
	return container.RegisterInstance(registry, "message", "ok")
}
