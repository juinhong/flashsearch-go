package engine

import "github.com/RoaringBitmap/roaring"

type TagSize struct {
	Name string
	Size uint64
}

type TagIndex struct {
	// Key: Tag Name (e.g., "color:red")
	// Value: Roaring Bitmap of Item IDs
	Tags map[string]*roaring.Bitmap
}
