package configdata

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultLocationsFor_whenWorkingDirDiffers_shouldIncludeResourceRootsInPriorityOrder(t *testing.T) {
	root := t.TempDir()
	executableDir := filepath.Join(root, "bin")
	workingDir := filepath.Join(root, "app")

	locations, err := defaultLocationsFor(executableDir, workingDir)
	if err != nil {
		t.Fatalf("default locations failed: %v", err)
	}

	want := []string{
		filepath.Clean(filepath.Join(executableDir, defaultResource)),
		filepath.Clean(filepath.Join(workingDir, defaultResource)),
	}
	if !reflect.DeepEqual(locations, want) {
		t.Fatalf("locations = %#v, want %#v", locations, want)
	}
}

func TestDefaultLocationsFor_whenResourcePathDuplicates_shouldDeduplicate(t *testing.T) {
	root := t.TempDir()

	locations, err := defaultLocationsFor(root, root)
	if err != nil {
		t.Fatalf("default locations failed: %v", err)
	}

	want := []string{
		filepath.Clean(filepath.Join(root, defaultResource)),
	}
	if !reflect.DeepEqual(locations, want) {
		t.Fatalf("locations = %#v, want %#v", locations, want)
	}
}
