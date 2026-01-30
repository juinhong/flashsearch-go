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

func TestAddProduct(t *testing.T) {
	idx := NewTagIndex()
	productID := uint32(101)
	tag := "color:blue"

	// This function doesn't exist yet!
	idx.Add(productID, tag)

	// Verify the ID was actually added
	if !idx.Contains(productID, tag) {
		t.Errorf("Expected ID %d to be found under tag %s", productID, tag)
	}
}

func TestContainProduct(t *testing.T) {
	idx := NewTagIndex()

	tag := "color:blue"

	productID := uint32(101)
	idx.Add(productID, tag)
	if !idx.Contains(productID, tag) {
		t.Errorf("Expected ID %d to be found under tag %s", productID, tag)
	}

	incorrectProductID := uint32(102)
	if idx.Contains(incorrectProductID, tag) {
		t.Errorf("ID %d shouldn't be found under tag %s", incorrectProductID, tag)
	}
}

func TestMultiAdd(t *testing.T) {
	idx := NewTagIndex()
	productID := uint32(500)
	tags := []string{
		"color:blue",
		"size:large",
		"brand:nike",
		"cat:shoes",
		"promo:active",
	}

	// Add the same ID to all 5 tags
	for _, tag := range tags {
		idx.Add(productID, tag)
	}

	// Assert the ID exists in every single bitmap
	for _, tag := range tags {
		if !idx.Contains(productID, tag) {
			t.Errorf("ID %d missing from tag %s", productID, tag)
		}
	}

	// Bonus: Check cardinality of each tag is exactly 1
	for _, tag := range tags {
		count := idx.Tags[tag].GetCardinality()
		if count != 1 {
			t.Errorf("Tag %s should have 1 item, but has %d", tag, count)
		}
	}
}
