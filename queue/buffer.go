package queue

import (
	"github.com/fasthall/gochariots/record"
)

type BufferHeap []record.Record

func (h BufferHeap) Len() int {
	return len(h)
}

func (h BufferHeap) Less(i, j int) bool {
	return h[i].TOId < h[j].TOId
}

func (h BufferHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *BufferHeap) Push(x interface{}) {
	*h = append(*h, x.(record.Record))
}

func (h *BufferHeap) Pop() interface{} {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[:n-1]
	return x
}
