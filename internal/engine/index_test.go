package engine

import "testing"

func TestNewIndex(t *testing.T) {
	// 1. Act: Try to create a new index
	idx := NewTagIndex()

	// 2. Assert: Check if it's nil
	if idx == nil {
		t.Fatal("NewTagIndex() returned nil, expected a valid pointer")
	}

	// 3. Assert: Check if the map is initialized
	if idx.Tags == nil {
		t.Error("Tag map should be initialized, not nil")
	}
}
