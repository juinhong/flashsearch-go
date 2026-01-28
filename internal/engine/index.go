package engine

import (
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
