package main

import (
	"flag"
	"flashsearch-go/internal/engine"
	"fmt"
	"log"
	"strings"
	"time"
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

	for i := uint32(0); i < 1000000; i++ {
		if i%2 == 0 {
			ti.Add(i, "tech")
		}

		if i%3 == 0 {
			ti.Add(i, "go")
		}
	}

	return ti
}
