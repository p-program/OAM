package set

import (
	"reflect"
	"testing"
)

func TestNewSortedMap_StringKeys(t *testing.T) {
	unsortedMap := map[string]int{
		"banana": 5,
		"apple":  2,
		"grape":  8,
		"pear":   3,
	}

	expectedKeys := []string{"apple", "banana", "grape", "pear"}

	sortedMap := NewSortedMap(unsortedMap)

	if !reflect.DeepEqual(sortedMap.SortedKeys, expectedKeys) {
		t.Errorf("expected keys %v, got %v", expectedKeys, sortedMap.SortedKeys)
	}

	for _, key := range expectedKeys {
		if sortedMap.Original[key] != unsortedMap[key] {
			t.Errorf("value mismatch for key %v: expected %v, got %v", key, unsortedMap[key], sortedMap.Original[key])
		}
	}
}

func TestNewSortedMap_IntKeys(t *testing.T) {
	unsortedMap := map[int]string{
		3: "three",
		1: "one",
		2: "two",
	}

	expectedKeys := []int{1, 2, 3}

	sortedMap := NewSortedMap(unsortedMap)

	if !reflect.DeepEqual(sortedMap.SortedKeys, expectedKeys) {
		t.Errorf("expected keys %v, got %v", expectedKeys, sortedMap.SortedKeys)
	}

	for _, key := range expectedKeys {
		if sortedMap.Original[key] != unsortedMap[key] {
			t.Errorf("value mismatch for key %v: expected %v, got %v", key, unsortedMap[key], sortedMap.Original[key])
		}
	}
}
