package engine

import (
	"container/heap"

	"github.com/RoaringBitmap/roaring"
)

type TagIndex struct {
	// Key: Tag Name (e.g., "color:red")
	// Value: Roaring Bitmap of Item IDs
	Tags map[string]*roaring.Bitmap
}

func NewTagIndex() *TagIndex {
	return &TagIndex{
		Tags: make(map[string]*roaring.Bitmap),
	}
}

// Add inserts a product ID into the bitmap for a specific tag
func (ti *TagIndex) Add(id uint32, tag string) {
	// If the tag doesn't exist yet, initialize a new bitmap
	if _, exists := ti.Tags[tag]; !exists {
		ti.Tags[tag] = roaring.New()
	}
	ti.Tags[tag].Add(id)
}

// Contains checks if an ID exists in a tag's bitmap
func (ti *TagIndex) Contains(id uint32, tag string) bool {
	bm, exists := ti.Tags[tag]
	if !exists || bm == nil {
		return false
	}

	return bm.Contains(id)
}

func (ti *TagIndex) AddMany(ids []uint32, tag string) {
	if _, exists := ti.Tags[tag]; !exists {
		ti.Tags[tag] = roaring.New()
	}

	ti.Tags[tag].AddMany(ids)
}

func (ti *TagIndex) SearchAND(tags ...string) *roaring.Bitmap {
	if len(tags) == 0 {
		return roaring.New()
	}

	// Start with a copy of the first tag's bitmap
	result, exists := ti.Tags[tags[0]]
	if !exists {
		result = roaring.New()
	}

	// We clone the first one so we don't modify the original index during the search
	finalResult := result.Clone()

	// Intersect with the remaining tags
	for i := 1; i < len(tags); i++ {
		other, exists := ti.Tags[tags[i]]
		if !exists {
			return roaring.New()
		}
		finalResult.And(other)
	}

	return finalResult
}

func (ti *TagIndex) Union(tags []string) *roaring.Bitmap {
	result := roaring.New()
	for _, tagName := range tags {
		if bm, exists := ti.Tags[tagName]; exists {
			result.Or(bm)
		}
	}

	return result
}

func (ti *TagIndex) Difference(tags []string) *roaring.Bitmap {
	includeBM, existsInclude := ti.Tags[tags[0]]
	excludeBM, existsExclude := ti.Tags[tags[1]]

	if !existsInclude {
		return roaring.New()
	}

	if !existsExclude {
		return includeBM.Clone()
	}

	// Create a copy so we don't modify the original 'include' tag
	result := includeBM.Clone()
	result.AndNot(excludeBM)

	return result
}

func (ti *TagIndex) Intersect(tags []string) *roaring.Bitmap {
	bmA, existsA := ti.Tags[tags[0]]
	bmB, existsB := ti.Tags[tags[1]]

	if !existsA || !existsB {
		return roaring.New()
	}

	result := bmA.Clone()
	result.And(bmB)

	return result
}

func (ti *TagIndex) FetchPage(bm *roaring.Bitmap, offset int, pageSize int) []uint32 {
	cardinality := int(bm.GetCardinality())
	if offset >= cardinality || pageSize <= 0 {
		return []uint32{}
	}

	it := bm.ManyIterator()

	// 1. "Burn" the offset
	// We use a small reusable buffer to discard the IDs before our page
	if offset > 0 {
		discardBuf := make([]uint32, 1024)
		for discarded := 0; discarded < offset; {
			toRead := offset - discarded
			if toRead > len(discardBuf) {
				toRead = len(discardBuf)
			}
			count := it.NextMany(discardBuf[:toRead])
			discarded += count
			if count == 0 {
				break
			}
		}
	}

	// 2. Collect the actual page
	results := make([]uint32, pageSize)
	actualCount := it.NextMany(results)

	return results[:actualCount]
}

func (ti *TagIndex) GetTopKTags(k int) []string {
	h := &TagHeap{}
	heap.Init(h)

	for name, bm := range ti.Tags {
		ts := TagSize{
			Name: name,
			Size: bm.GetCardinality(),
		}

		if h.Len() < k {
			heap.Push(h, ts)
		} else if ts.Size > (*h)[0].Size {
			// If current tag is bigger than the "smallest of the best"
			heap.Pop(h)
			heap.Push(h, ts)
		}
	}

	// Extract and reverse (since Pop gives smallest first)
	result := make([]string, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(TagSize).Name
	}

	return result
}
