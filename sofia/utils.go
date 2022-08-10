package sofia

// Sequence
type Sequence struct {
	pool []bool // Raw indices
}

// Create new sequence
func NewSequence(size uint8) *Sequence {
	// Allocate new sequence
	seq := new(Sequence)

	// Make pool
	seq.pool = make([]bool, size)

	return seq
}

// Delete a sequence
func DeleteSequence(seq *Sequence) {
	// GC
}

// Get index
func (seq *Sequence) GetIndex() uint8 {
	for idx := 0; idx < len(seq.pool); idx++ {
		if !seq.pool[idx] {
			seq.pool[idx] = true
			return uint8(idx + 1)
		}
	}

	return 0
}

// Free index
func (seq *Sequence) FreeIndex(idx uint8) {
	if idx != 0 {
		seq.pool[idx-1] = false
	}
}
