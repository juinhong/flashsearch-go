package engine

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/RoaringBitmap/roaring"
)

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

func TestRunOptimization(t *testing.T) {
	idx := NewTagIndex()
	tag := "bulk-import"

	// 1. Add 1,000 sequential IDs.
	// Initially, Roaring usually stores these in an Array or Bitmap Container.
	for i := uint32(1); i <= 1000; i++ {
		idx.Add(i, tag)
	}

	bm := idx.Tags[tag]

	// Get size before optimization
	sizeBefore := bm.GetSerializedSizeInBytes()

	// 2. Perform the Magic
	// This scans the bitmap and converts sequences into RunContainers (Start/Length)
	bm.RunOptimize()

	// Get stats after optimization
	statsAfter := bm.Stats()

	// Get size after
	sizeAfter := bm.GetSerializedSizeInBytes()

	// 3. Assertion: RunContainers should increase, and overall size should decrease
	if statsAfter.RunContainers == 0 {
		t.Error("RunOptimize failed to create a RunContainer for sequential data")
	}

	if sizeAfter >= sizeBefore {
		t.Errorf("Optimization didn't reduce size. Before: %d, After: %d",
			sizeBefore, sizeAfter)
	}
}

func TestSerialization(t *testing.T) {
	idx := NewTagIndex()
	tag := "persistent-tag"

	// Add some data to ensure we have something to serialize
	idx.Add(1, tag)
	idx.Add(100, tag)
	idx.Add(65536, tag) // This forces at least two containers

	bm := idx.Tags[tag]

	// Use a bytes.Buffer as our "disk"
	var buf bytes.Buffer

	// WriteTo writes the bitmap in the standard Roaring format
	n, err := bm.WriteTo(&buf)

	if err != nil {
		t.Errorf("Failed to serialize bitmap: %v", err)
	}

	if n == 0 || buf.Len() == 0 {
		t.Error("Serialized buffer is empty")
	}

	// Assertion: Ensure the serialized size matches the method we used yesterday
	if int64(buf.Len()) != int64(bm.GetSerializedSizeInBytes()) {
		t.Errorf("Buffer length %d doesn't match expected size %d",
			buf.Len(), bm.GetSerializedSizeInBytes())
	}
}

func TestDeserialization(t *testing.T) {
	// 1. Setup original data
	original := roaring.New()
	original.Add(1)
	original.Add(100)
	original.Add(70000) // Puts data in Container 0 and Container 1

	// 2. Serialize to a buffer
	var buf bytes.Buffer
	_, err := original.WriteTo(&buf)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// 3. Deserialize into a NEW object
	restored := roaring.New()
	_, err = restored.ReadFrom(&buf)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	// 4. Assertions
	if !original.Equals(restored) {
		t.Error("Restored bitmap does not match original")
	}

	if restored.GetCardinality() != 3 {
		t.Errorf("Expected cardinality 3, got %d", restored.GetCardinality())
	}
}

func BenchmarkRoaringIntersection(b *testing.B) {
	// 1. Setup: Create two dense bitmaps with 1M items each
	bm1 := roaring.New()
	bm2 := roaring.New()

	for i := uint32(0); i < 1000000; i++ {
		bm1.Add(i)
		// Offset bm2 slightly so the intersection actually has work to do
		bm2.Add(i + 500000)
	}

	// Run the actual benchmark
	b.ResetTimer() // Don't count the setup time!
	for i := 0; i < b.N; i++ {
		_ = roaring.And(bm1, bm2)
	}
}

func BenchmarkMapIntersection(b *testing.B) {
	// 1. Setup: Two maps with 1M items each
	map1 := make(map[uint32]bool)
	map2 := make(map[uint32]bool)

	for i := uint32(0); i < 1000000; i++ {
		map1[i] = true
		map2[i+500000] = true // Offset by 50%
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Manual intersection
		result := make(map[uint32]bool)
		// Smallest-to-largest optimization (manual)
		if len(map1) < len(map2) {
			for k := range map1 {
				if map2[k] {
					result[k] = true
				}
			}
		} else {
			for k := range map2 {
				if map1[k] {
					result[k] = true
				}
			}
		}
	}
}

func TestFetchPage(t *testing.T) {
	ti := NewTagIndex()
	bm := roaring.New()
	// Add IDs 0, 10, 20, 30, 40, 50
	for i := uint32(0); i <= 5; i++ {
		bm.Add(i * 10)
	}

	tests := []struct {
		name     string
		offset   int
		pageSize int
		want     []uint32
	}{
		{"First Page", 0, 2, []uint32{0, 10}},
		{"Second Page", 2, 2, []uint32{20, 30}},
		{"Last Partial Page", 4, 2, []uint32{40, 50}},
		{"Out of Bounds", 10, 2, []uint32{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ti.FetchPage(bm, tt.offset, tt.pageSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchPage(%d, %d) = %v; want %v", tt.offset, tt.pageSize, got, tt.want)
			}
		})
	}
}
