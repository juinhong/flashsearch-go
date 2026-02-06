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

func TestBulkLoad(t *testing.T) {
	idx := NewTagIndex()
	tag := "category:electronics"

	ids := make([]uint32, 10000)
	for i := 0; i < 10000; i++ {
		ids[i] = uint32(i)
	}

	idx.AddMany(ids, tag)

	count := idx.Tags[tag].GetCardinality()
	if count != 10000 {
		t.Errorf("Expected 10,000 IDs, but got %d", count)
	}
}

func TestIntersection(t *testing.T) {
	idx := NewTagIndex()

	// Setup: Red items (1, 2, 3)
	idx.Add(1, "color:red")
	idx.Add(2, "color:red")
	idx.Add(3, "color:red")

	// Setup: Large items (2, 3, 4)
	idx.Add(2, "size:large")
	idx.Add(3, "size:large")
	idx.Add(4, "size:large")

	// The logic we need to build:
	results := idx.SearchAND("color:red", "size:large")

	// Assertions: Only IDs 2 and 3 should be returned
	if results.GetCardinality() != 2 {
		t.Errorf("Expected 2 results, got %d", results.GetCardinality())
	}

	if !results.Contains(2) || !results.Contains(3) {
		t.Error("Result set missing ID 2 or 3")
	}

	if results.Contains(1) || results.Contains(4) {
		t.Error("Result set contains IDs that do not match both tags")
	}
}

func TestUnion(t *testing.T) {
	idx := NewTagIndex()

	// Setup: Red has IDs 1, 2. Blue has IDs 2, 3.
	idx.Add(1, "color:red")
	idx.Add(2, "color:red")
	idx.Add(2, "color:blue")
	idx.Add(3, "color:blue")

	// We want the union of Red and Blue
	tags := []string{"color:red", "color:blue"}
	result := idx.Union(tags)

	// The result should have IDs 1, 2, 3 (no duplicates)
	if result.GetCardinality() != 3 {
		t.Errorf("Expected cardinality 3, got %d", result.GetCardinality())
	}

	expected := []uint32{1, 2, 3}
	for _, id := range expected {
		if !result.Contains(id) {
			t.Errorf("Expected result to contain ID %d", id)
		}
	}
}

func TestDifference(t *testing.T) {
	idx := NewTagIndex()

	// Setup: IDs 1, 2, 3 are Red. IDs 2, 4 are on Sale.
	idx.Add(1, "color:red")
	idx.Add(2, "color:red")
	idx.Add(3, "color:red")

	idx.Add(2, "status:sale")
	idx.Add(4, "status:sale")

	// Logic: Red AND NOT Sale
	tags := []string{"color:red", "status:sale"}
	result := idx.Difference(tags)

	// ID 2 should be removed because it is on Sale.
	// Result should be [1, 3]
	if result.GetCardinality() != 2 {
		t.Errorf("Expected cardinality 2, got %d", result.GetCardinality())
	}

	if result.Contains(2) {
		t.Error("Result should NOT contain ID 2 (it's on sale)")
	}
}

func TestCompound(t *testing.T) {
	idx := NewTagIndex()

	// Tag A (Nike): IDs 1, 2
	idx.Add(1, "brand:nike")
	idx.Add(2, "brand:nike")

	// Tag B (Adidas): IDs 3, 4
	idx.Add(3, "brand:adidas")
	idx.Add(4, "brand:adidas")

	// Tag C (On Sale): IDs 2, 4
	idx.Add(2, "status:sale")
	idx.Add(4, "status:sale")

	// Goal: (Nike OR Adidas) AND Sale
	// 1. Union (Nike, Adidas) -> {1, 2, 3, 4}
	// 2. Intersect with Sale -> {2, 4}

	tags := []string{"brand:nike", "brand:adidas"}
	brands := idx.Union(tags)
	saleBM, _ := idx.Tags["status:sale"]

	// Intersect the result of the Union with the Sale bitmap
	brands.And(saleBM)

	if brands.GetCardinality() != 2 {
		t.Errorf("Expected 2 items, got %d", brands.GetCardinality())
	}

	if !brands.Contains(2) || !brands.Contains(4) {
		t.Errorf("Result missing expected IDs 2 or 4")
	}
}

func TestEmptyResults(t *testing.T) {
	idx := NewTagIndex()

	// Add some data so the index isn't totally empty
	idx.Add(1, "color:red")

	// 1. Test Intersect with a non-existent tag
	// "color:red" AND "brand:apple" (which doesn't exist)
	tags := []string{"color:red", "brand:apple"}
	result := idx.Intersect(tags)

	if result == nil {
		t.Fatal("Intersect returned nil; expected an empty bitmap")
	}

	if result.GetCardinality() != 0 {
		t.Errorf("Expected 0 results, got %d", result.GetCardinality())
	}

	// 2. Test Union with only non-existent tags
	tags = []string{"fake:1", "fake:2"}
	resultUnion := idx.Union(tags)
	if resultUnion == nil {
		t.Fatal("Union returned nil; expected an empty bitmap")
	}
}
