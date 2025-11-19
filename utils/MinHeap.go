package utils


type Node struct {
	Symbol byte
	Freq   int
	Left   *Node
	Right  *Node
}

type MinHeap []*Node

func (heap MinHeap) Len() int { return len(heap) }
func (heap MinHeap) less(i, j int) bool { return heap[i].Freq < heap[j].Freq }
func (heap MinHeap) swap(i, j int) { heap[i], heap[j] = heap[j], heap[i] }
func parent(i int) int { return (i - 1) / 2 }
func left(i int) int { return 2*i + 1 }
func right(i int) int { return 2*i + 2 }


func (heap *MinHeap) Up(i int) {
	equilibrium := false
	for i > 0 && !equilibrium {
		p := parent(i)
		if !heap.less(i, p) {
			equilibrium = true
		} else {
			heap.swap(i, p)
			i = p
		}
	}
}

func (heap *MinHeap) Down(i int) {
	equilibrium := false
	n := heap.Len()
	for i < n && !equilibrium {
		l := left(i)
		r := right(i)
		smallest := i
		if l < n && heap.less(l, smallest) {
			smallest = l
		}
		if r < n && heap.less(r, smallest) {
			smallest = r
		}
		if smallest == i {
			equilibrium = true
		} else {
			heap.swap(i, smallest)
			i = smallest
		}
	}
}

func (heap *MinHeap) Insert(node *Node) {
	*heap = append(*heap, node)
	heap.Up(heap.Len() - 1)
}

func (heap *MinHeap) Pop() *Node {
	n := heap.Len()
	if n == 0 {
		return &Node{}
	}
	min := (*heap)[0]
	(*heap)[0] = (*heap)[n-1]
	*heap = (*heap)[:n-1]
	heap.Down(0)
	return min
}

func BuildHeap(data []byte) MinHeap {
	freqTable := make(map[byte]int)

	for _, b := range data {
		freqTable[b]++
	}

	heap := MinHeap{}
	for b, freq := range freqTable {
		heap.Insert(&Node{Symbol: b, Freq: freq, Left: nil, Right: nil})
	}

	return heap
}