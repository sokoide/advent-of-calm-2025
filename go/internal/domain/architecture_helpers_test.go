package domain

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	t.Run("should merge multiple maps", func(t *testing.T) {
		m1 := map[string]any{"key1": "val1"}
		m2 := map[string]any{"key2": "val2"}
		m3 := map[string]any{"key3": "val3"}

		expected := map[string]any{
			"key1": "val1",
			"key2": "val2",
			"key3": "val3",
		}

		result := Merge(m1, m2, m3)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("should panic on key collision", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic on key collision")
			}
		}()

		m1 := map[string]any{"key1": "val1"}
		m2 := map[string]any{"key1": "val2"} // Collision

		Merge(m1, m2)
	})

	t.Run("should handle empty input", func(t *testing.T) {
		result := Merge()
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})
}
