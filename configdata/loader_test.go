package configdata_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"goark.dev/boot/configdata"
)

func TestLoad_whenDefaultAndProfileFilesExist_shouldApplySpringLikePriority(t *testing.T) {
	root := t.TempDir()
	conf := filepath.Join(root, "conf")
	mkdir(t, conf)
	writeFile(t, filepath.Join(conf, "app.yml"), `
spring:
  profiles:
    active: dev
server:
  port: 8080
hosts:
  - conf
  - base
`)
	writeFile(t, filepath.Join(root, "app.yml"), `
server:
  port: 8081
app:
  name: root
`)
	writeFile(t, filepath.Join(conf, "app-dev.toml"), `
[server]
port = 9090
[feature]
enabled = true
`)
	writeFile(t, filepath.Join(root, "app-dev.properties"), `
server.port=9091
server.host=127.0.0.1
`)

	result, err := configdata.Load(context.Background(), configdata.WithLocations(conf, root))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if !reflect.DeepEqual(result.Profiles, []string{"dev"}) {
		t.Fatalf("unexpected profiles: %#v", result.Profiles)
	}
	if got := mustGet(t, result, "server.port"); got != "9091" {
		t.Fatalf("expected executable profile to win, got %q", got)
	}
	if got := mustGet(t, result, "server.host"); got != "127.0.0.1" {
		t.Fatalf("unexpected server host: %q", got)
	}
	if got := mustGet(t, result, "feature.enabled"); got != "true" {
		t.Fatalf("unexpected feature flag: %q", got)
	}
	if got := mustGet(t, result, "app.name"); got != "root" {
		t.Fatalf("unexpected app name: %q", got)
	}
	if got := mustGet(t, result, "hosts"); got != "conf,base" {
		t.Fatalf("unexpected list flattening: %q", got)
	}
	if len(result.Sources) != 4 {
		t.Fatalf("expected four loaded sources, got %#v", result.Sources)
	}
	if result.Sources[0].Location != filepath.Clean(conf) || result.Sources[3].Location != filepath.Clean(root) {
		t.Fatalf("unexpected source order: %#v", result.Sources)
	}
}

func TestLoad_whenProfilesAreExplicit_shouldIgnoreProfileFromBaseConfig(t *testing.T) {
	root := t.TempDir()
	conf := filepath.Join(root, "conf")
	mkdir(t, conf)
	writeFile(t, filepath.Join(conf, "app.yml"), `
spring:
  profiles:
    active: dev
server:
  port: 8080
`)
	writeFile(t, filepath.Join(conf, "app-prod.yml"), `
server:
  port: 9090
`)

	result, err := configdata.Load(context.Background(), configdata.WithLocations(conf, root), configdata.WithProfiles("prod"))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if !reflect.DeepEqual(result.Profiles, []string{"prod"}) {
		t.Fatalf("unexpected profiles: %#v", result.Profiles)
	}
	if got := mustGet(t, result, "server.port"); got != "9090" {
		t.Fatalf("expected prod profile value, got %q", got)
	}
}

func TestLoad_whenMultipleFormatsHaveSameName_shouldUseFormatPriority(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
app:
  name: yaml
`)
	writeFile(t, filepath.Join(root, "app.toml"), `
[app]
name = "toml"
`)
	writeFile(t, filepath.Join(root, "app.properties"), "app.name=properties\n")

	result, err := configdata.Load(context.Background(), configdata.WithLocations(root))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if got := mustGet(t, result, "app.name"); got != "yaml" {
		t.Fatalf("expected yaml to win, got %q", got)
	}
	if len(result.Sources) != 1 || result.Sources[0].Format != configdata.FormatYAML {
		t.Fatalf("unexpected loaded sources: %#v", result.Sources)
	}
}

func TestLoad_whenArgsSpecifyConfigFileAndProfile_shouldLoadExactFileAndProfileVariant(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "custom.yaml")
	writeFile(t, base, `
app:
  name: base
feature:
  enabled: false
`)
	writeFile(t, filepath.Join(root, "custom-dev.yaml"), `
app:
  name: dev
feature:
  enabled: true
`)

	result, err := configdata.Load(
		context.Background(),
		configdata.WithArgs(
			"--spring.config.location", base,
			"--spring.profiles.active=dev",
		),
	)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if got := mustGet(t, result, "app.name"); got != "dev" {
		t.Fatalf("expected profile config to win, got %q", got)
	}
	if got := mustGet(t, result, "feature.enabled"); got != "true" {
		t.Fatalf("expected profile feature flag, got %q", got)
	}
	if !reflect.DeepEqual(result.Profiles, []string{"dev"}) {
		t.Fatalf("profiles = %#v, want dev", result.Profiles)
	}
	if !reflect.DeepEqual(result.Environment.ActiveProfiles(), []string{"dev"}) {
		t.Fatalf("environment profiles = %#v, want dev", result.Environment.ActiveProfiles())
	}
	if len(result.Sources) != 2 ||
		result.Sources[0].Path != filepath.Clean(base) ||
		result.Sources[1].Path != filepath.Clean(filepath.Join(root, "custom-dev.yaml")) {
		t.Fatalf("unexpected sources: %#v", result.Sources)
	}
}

func TestLoad_whenEnvironmentSpecifiesLocationNameAndProfile_shouldApplyEnvironment(t *testing.T) {
	root := t.TempDir()
	t.Setenv(configdata.EnvSpringConfigLocation, root)
	t.Setenv(configdata.EnvSpringConfigName, "service")
	t.Setenv(configdata.EnvSpringProfilesActive, "prod")
	writeFile(t, filepath.Join(root, "service.properties"), `
app.name=base
`)
	writeFile(t, filepath.Join(root, "service-prod.properties"), `
app.name=prod
`)

	result, err := configdata.Load(context.Background())
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if got := mustGet(t, result, "app.name"); got != "prod" {
		t.Fatalf("expected env profile config to win, got %q", got)
	}
	if !reflect.DeepEqual(result.Profiles, []string{"prod"}) {
		t.Fatalf("profiles = %#v, want prod", result.Profiles)
	}
	if len(result.Sources) != 2 {
		t.Fatalf("sources = %#v, want base and profile", result.Sources)
	}
}

func TestLoad_whenNoConfigFilesExist_shouldReturnBuiltInDefaults(t *testing.T) {
	result, err := configdata.Load(context.Background(), configdata.WithLocations(t.TempDir()))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if len(result.Sources) != 1 || !result.Sources[0].BuiltIn {
		t.Fatalf("expected built-in source, got %#v", result.Sources)
	}
	if got := mustGet(t, result, configdata.PropertySpringApplicationName); got != "goark" {
		t.Fatalf("built-in application name = %q", got)
	}
	if got := mustGet(t, result, configdata.PropertySpringConfigName); got != "app" {
		t.Fatalf("built-in config name = %q, want app", got)
	}
}

func TestLoad_whenDefaultExists_shouldLoadApp(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), "app:\n  name: goark\n")

	result, err := configdata.Load(context.Background(), configdata.WithLocations(root))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if got := mustGet(t, result, "app.name"); got != "goark" {
		t.Fatalf("app.name = %q, want goark", got)
	}
}

func TestLoad_whenPropertySourcesConflict_shouldApplyBootPriority(t *testing.T) {
	root := t.TempDir()
	t.Setenv("SERVER_PORT", "9090")
	writeFile(t, filepath.Join(root, "app.yml"), `
spring:
  profiles:
    active: dev
server:
  port: 8080
`)
	writeFile(t, filepath.Join(root, "app-dev.yml"), "server:\n  port: 8081\n")

	result, err := configdata.Load(
		context.Background(),
		configdata.WithLocations(root),
		configdata.WithArgs("--server.port=10090"),
	)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if got := mustGet(t, result, "server.port"); got != "10090" {
		t.Fatalf("server.port = %q, want command-line value", got)
	}

	result, err = configdata.Load(context.Background(), configdata.WithLocations(root), configdata.WithArgs())
	if err != nil {
		t.Fatalf("load config without explicit args failed: %v", err)
	}
	if got := mustGet(t, result, "server.port"); got != "9090" {
		t.Fatalf("server.port = %q, want environment value", got)
	}
}

func TestLoad_whenAdditionalLocationProvided_shouldOverrideDefaultLocation(t *testing.T) {
	base := t.TempDir()
	additional := t.TempDir()
	writeFile(t, filepath.Join(base, "app.yml"), "app:\n  name: base\n")
	writeFile(t, filepath.Join(additional, "app.yml"), "app:\n  name: additional\n")

	result, err := configdata.Load(
		context.Background(),
		configdata.WithLocations(base),
		configdata.WithArgs("--spring.config.additional-location="+additional),
	)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if got := mustGet(t, result, "app.name"); got != "additional" {
		t.Fatalf("app.name = %q, want additional", got)
	}
}

func TestLoad_whenProfilesUseIncludeDefaultAndGroup_shouldResolveProfileSet(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
spring:
  profiles:
    default: local
    include: audit
    group:
      local: web,data
`)
	for _, profile := range []string{"local", "web", "data", "audit"} {
		writeFile(t, filepath.Join(root, "app-"+profile+".yml"), "loaded:\n  "+profile+": true\n")
	}

	result, err := configdata.Load(context.Background(), configdata.WithLocations(root))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	want := []string{"local", "web", "data", "audit"}
	if !reflect.DeepEqual(result.Profiles, want) {
		t.Fatalf("profiles = %#v, want %#v", result.Profiles, want)
	}
	for _, profile := range want {
		if got := mustGet(t, result, "loaded."+profile); got != "true" {
			t.Fatalf("loaded.%s = %q, want true", profile, got)
		}
	}
}

func TestLoad_whenProfileGroupsAreCircular_shouldReject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
spring:
  profiles:
    active: a
    group:
      a: b
      b: a
`)

	if _, err := configdata.Load(context.Background(), configdata.WithLocations(root)); err == nil {
		t.Fatal("load config should reject circular profile groups")
	}
}

func mustGet(t *testing.T, result *configdata.Result, key string) string {
	t.Helper()
	value, ok := result.Environment.GetProperty(key)
	if !ok {
		t.Fatalf("expected key %q", key)
	}
	return value
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %q failed: %v", path, err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q failed: %v", path, err)
	}
}
