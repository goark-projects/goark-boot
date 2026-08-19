package configdata_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/goark-projects/boot/configdata"
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

func TestLoad_whenNoConfigFilesExist_shouldReturnEmptyEnvironment(t *testing.T) {
	result, err := configdata.Load(context.Background(), configdata.WithLocations(t.TempDir()))
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if len(result.Sources) != 0 {
		t.Fatalf("expected no sources, got %#v", result.Sources)
	}
	if len(result.Environment.Keys()) != 0 {
		t.Fatalf("expected empty environment, got %#v", result.Environment.Snapshot())
	}
}

func mustGet(t *testing.T, result *configdata.Result, key string) string {
	t.Helper()
	value, ok := result.Environment.Get(key)
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
