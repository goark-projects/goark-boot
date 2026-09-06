package configdata_test

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"goark.dev/boot/configdata"
)

func TestLoad_whenProfilesUseIncludeDefaultAndGroup_shouldResolveProfileSet(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "app.yml"), `
goark:
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
goark:
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
