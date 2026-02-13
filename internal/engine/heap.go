package engine

// TagHeap implements heap.Interface and holds TagSizes.
// We use a MIN-HEAP so that the smallest of the "Top K" is at the root.
type TagHeap []TagSize

func (h TagHeap) Len() int {
	return len(h)
}

func (h TagHeap) Less(i, j int) bool {
	return h[i].Size < h[j].Size
}

func (h TagHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *TagHeap) Push(x interface{}) {
	*h = append(*h, x.(TagSize))
}

func (h *TagHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
