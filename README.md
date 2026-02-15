# FlashSearch-Go ⚡

A high-performance, in-memory search engine built on **Roaring Bitmaps**. This engine is designed for hardware-sympathetic tag indexing, allowing for complex intersections across millions of records with sub-millisecond latency.

## 🚀 Key Features
- **Microsecond Intersections:** Perform `AND`/`OR` operations on millions of IDs in < 1ms.
- **Adaptive Compression:** Utilizes Array, Bitmap, and Run-Length Encoding (RLE) containers to optimize memory based on data density.
- **S2L Optimization:** Smallest-to-Largest query planning to minimize CPU cycles during multi-tag intersections.
- **Efficient Pagination:** Custom "Burn-logic" pagination using `ManyIterator` for high-throughput result streaming.
- **Top-K Analytics:** $O(N \log K)$ ranking using a Min-Heap to find the most popular tags without sorting the entire index.

## 🏗️ Architecture
The engine is structured as a "Compressed Inverted Index." Instead of storing lists of IDs, we store bit-compressed containers.

### How Roaring Works
The ID space is divided into chunks of $2^{16}$ (65,536) integers. Depending on how many IDs are in a chunk, the engine chooses the most efficient storage format:
1. **Array Containers:** For sparse data (< 4,096 IDs).
2. **Bitmap Containers:** For dense data (fixed 8KB bitset).
3. **Run Containers:** For contiguous sequences (e.g., IDs 1000-5000) using RLE.



## 📂 Project Structure
```text
flashsearch-go/
├── cmd/
│   └── flashsearch/       # CLI Entry Point
├── internal/
│   └── engine/            # Core Bitmap Logic
│       ├── heap.go        # Top-K Min-Heap implementation
│       ├── index.go       # Search & Intersection logic
│       ├── model.go       # Data structures
│       └── util.go        # Iterators & Pagination helpers
└── go.mod
```

## 🛠️ Getting Started
1. Installation
```
git clone [https://github.com/juinhong/flashsearch-go](https://github.com/juinhong/flashsearch-go)
cd flashsearch-go
go mod tidy
```

## Running the CLI
Run a search across multiple tags:
```
go run ./cmd/flashsearch -tags=golang,fast,database -limit=10
```

## Running Benchmarks
To see the hardware efficiency in action:
```
go test ./internal/engine/... -bench=. -benchmem
```