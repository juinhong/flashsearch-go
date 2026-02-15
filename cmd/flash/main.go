package main

import (
	"flag"
	"flashsearch-go/internal/engine"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
)

func main() {
	// 1. Define CLI Flags
	tagsFlag := flag.String("tags", "", "Comma-separated list of tags to intersect")
	limit := flag.Int("limit", 10, "Number of results to show")
	flag.Parse()

	if *tagsFlag == "" {
		log.Fatal("Please provide tags, e.g., -tags=tech,go")
	}

	// 2. Setup Dummy Data (1M records)
	index := setupData()

	// 3. Process Query
	searchTags := strings.Split(*tagsFlag, ",")
	fmt.Printf("🔍 Searching for: %v\n", searchTags)

	start := time.Now()
	result := index.IntersectMany(searchTags)
	duration := time.Since(start)

	// 4. Display Results
	count := result.GetCardinality()
	fmt.Printf("✅ Found %d matches in %v\n", count, duration)

	if count > 0 {
		fmt.Printf("📄 Top %d IDs: %v\n", *limit, index.FetchPage(result, 0, *limit))
	}
}

func setupData() *engine.TagIndex {
	ti := engine.NewTagIndex()
	// Create tags with different distribution patterns
	ti.Tags["all"] = roaring.New()  // 5,000,000 IDs
	ti.Tags["even"] = roaring.New() // 2,500,000 IDs
	ti.Tags["tri"] = roaring.New()  // 1,666,666 IDs
	ti.Tags["rare"] = roaring.New() // 1 ID

	fmt.Println("🏗️  Generating 5,000,000 IDs...")
	for i := uint32(0); i < 5000000; i++ {
		ti.Tags["all"].Add(i)
		if i%2 == 0 {
			ti.Tags["even"].Add(i)
		}
		if i%3 == 0 {
			ti.Tags["tri"].Add(i)
		}
	}
	ti.Tags["rare"].Add(4999999)

	// Optimize the bitmaps for the best performance
	for _, bm := range ti.Tags {
		bm.RunOptimize() // Crucial for RLE compression
	}
	return ti
}
