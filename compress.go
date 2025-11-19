package main

import (
	"fmt"
	"os"
	"encoding/binary"
	"kryptr/utils"
)

func buildHuffmanTree(heap utils.MinHeap) *utils.Node {
	for heap.Len() > 1 {
		left := heap.Pop()
		right := heap.Pop()

		merged := utils.Node{
			Symbol:  0,
			Freq:   left.Freq + right.Freq,
			Left:   left,
			Right:  right,
		}
		heap.Insert(&merged)
	}
	return heap.Pop()
}

func serializeTree(root *utils.Node) []byte {
    var out []byte
    walk(root, &out)
    return out
}

func walk (n *utils.Node, out *[]byte)  {
	    if n.Left == nil && n.Right == nil {
            *out = append(*out, 1)
            *out = append(*out, n.Symbol)
            return
        }

        *out = append(*out, 0)
        walk(n.Left, out)
        walk(n.Right, out)
		return
}

func DeserializeToDict(data []byte) (map[string]byte, int) {
    dict := make(map[string]byte)
    i := 0
    buildDict(data, &i, "", dict)
    return dict, i
}

func buildDict(data []byte, index *int, prefix string, dict map[string]byte) {
    marker := data[*index]
    *index++

    if marker == 1 {
        symbol := data[*index]
        *index++
        dict[prefix] = symbol
        return
    }

    // Internal node → descend left as 0, right as 1
    buildDict(data, index, prefix+"0", dict)
    buildDict(data, index, prefix+"1", dict)
}

func createCompressionDictionary(tree *utils.Node, prefix string, compressionDict map[byte]string) {
	if tree.Left == nil && tree.Right == nil {
		compressionDict[tree.Symbol] = prefix
		return
	}
	if tree.Left != nil {
		createCompressionDictionary(tree.Left, prefix+"0", compressionDict)
	}
	if tree.Right != nil {
		createCompressionDictionary(tree.Right, prefix+"1", compressionDict)
	}
}

func huffmanCompress(data []byte) []byte {
	heap := utils.BuildHeap(data)
	huffManTree := buildHuffmanTree(heap)
	compressionDict := make(map[byte]string)
	createCompressionDictionary(huffManTree, "", compressionDict)

	serializedTree := serializeTree(huffManTree)

	var bw utils.BitWriter

	lengthBits := fmt.Sprintf("%032b", len(data))
	bw.WriteBits(lengthBits)

    for _, b := range data {
        bw.WriteBits(compressionDict[b])
    }

    compressed := bw.Finalize()
	compressed = append(serializedTree, compressed...)

	return compressed
}

func huffmanDecompress(packed []byte) []byte {
    var out []byte
    var current string

	i := 0
	dict, i := DeserializeToDict(packed)
	compressedData := packed[i:]

	lenMessage := binary.BigEndian.Uint32(compressedData[:4])
	symbolsRead := 0
	fmt.Println("Length of message:", lenMessage)

    for _, b := range compressedData[4:] {
        for i := 7; i >= 0; i-- {
            bit := (b >> i) & 1
            if bit == 1 {
                current += "1"
            } else {
                current += "0"
            }

            if val, ok := dict[current]; ok {
                out = append(out, val)
                current = ""
				symbolsRead++
            }

			if symbolsRead >= int(lenMessage) {
				return out
			}
        }
    }

    return out
}

func main() {
	data := []byte("Sample data for compression")
	if b, err := os.ReadFile("test.txt"); err == nil {
		data = b
	} else {
		fmt.Println("warning: could not read test.txt:", err)
	}
	fmt.Println("Original size in bytes:", len(data))
	bitsLen := len(data) * 8
	fmt.Println("Length in bits:", bitsLen)
	compressedData := huffmanCompress(data)

	decompressedData := huffmanDecompress(compressedData) // You need to pass the correct dictionary here
	bitsLen = len(compressedData) * 8
	fmt.Println("Length in bits compressed:", bitsLen)
	fmt.Println("Compression ratio:", float64(len(compressedData)*8)/float64(len(decompressedData)*8))
}