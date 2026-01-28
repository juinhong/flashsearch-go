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
