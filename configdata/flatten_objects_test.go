package configdata

import "testing"

func TestFlattenObjectListsPreservesStructuredValues(t *testing.T) {
	values := flattenSettings(map[string]any{"groups": []any{map[string]any{"name": "admin", "paths": []any{"/admin/**"}}}, "simple": []any{"a", "b"}})
	if values["simple"] != "a,b" {
		t.Fatal("标量列表行为发生变化")
	}
	if values["groups[0].paths"] != "/admin/**" || values["groups[0].name"] != "admin" {
		t.Fatal(values)
	}
}
