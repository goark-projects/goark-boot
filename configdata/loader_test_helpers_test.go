package configdata_test

import (
	"os"
	"testing"

	"goark.dev/boot/configdata"
)

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
