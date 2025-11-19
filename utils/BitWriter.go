package utils

type BitWriter struct {
    buf  []byte
    curr byte
    n    int // number of bits written into curr (0..7)
}

func (w *BitWriter) WriteBit(bit byte) {
    w.curr <<= 1
    w.curr |= bit & 1
    w.n++

    if w.n == 8 {
        w.buf = append(w.buf, w.curr)
        w.curr = 0
        w.n = 0
    }
}

func (w *BitWriter) WriteBits(code string) {
    for i := 0; i < len(code); i++ {
        if code[i] == '0' {
            w.WriteBit(0)
        } else {
            w.WriteBit(1)
        }
    }
}

func (w *BitWriter) Finalize() []byte {
    if w.n > 0 {
        w.curr <<= uint(8 - w.n)
        w.buf = append(w.buf, w.curr)
    }
    return w.buf
}