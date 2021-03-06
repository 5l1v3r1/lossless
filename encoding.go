package lossless

func encodeByte(w *bitWriter, p ByteProbs, b byte) error {
	evt := fullEvent(p)
	for evt.Len() > 1 {
		zero, one := evt.Split()
		var err error
		if zero.Contains(b) {
			evt = zero
			err = w.WriteBit(false)
		} else {
			evt = one
			err = w.WriteBit(true)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeByte(r *bitReader, p ByteProbs) (byte, error) {
	evt := fullEvent(p)
	for evt.Len() > 1 {
		zero, one := evt.Split()
		b, err := r.ReadBit()
		if err != nil {
			return 0, err
		}
		if b {
			evt = one
		} else {
			evt = zero
		}
	}
	return evt.bytes[0], nil
}

type event struct {
	bytes []byte
	probs []float64
}

func fullEvent(probs ByteProbs) *event {
	res := &event{
		bytes: make([]byte, len(probs)),
		probs: make([]float64, len(probs)),
	}
	for i, x := range probs {
		res.bytes[i] = byte(i)
		res.probs[i] = x
	}
	res.arrangePoles()
	return res
}

func (e *event) Len() int {
	return len(e.bytes)
}

func (e *event) Swap(i, j int) {
	e.bytes[i], e.bytes[j] = e.bytes[j], e.bytes[i]
	e.probs[i], e.probs[j] = e.probs[j], e.probs[i]
}

func (e *event) Subset(start, end int) *event {
	return &event{
		bytes: e.bytes[start:end],
		probs: e.probs[start:end],
	}
}

func (e *event) Contains(b byte) bool {
	for _, b1 := range e.bytes {
		if b1 == b {
			return true
		}
	}
	return false
}

func (e *event) Split() (e1, e2 *event) {
	if e.Len() == 0 {
		return e, e
	} else if e.probs[0] == 0 {
		return e.Subset(0, e.Len()/2), e.Subset(e.Len()/2, e.Len())
	}

	var probSum float64
	halfSum := e.probSum() * 0.5
	for i, p := range e.probs {
		if probSum >= halfSum {
			return e.Subset(0, i), e.Subset(i, e.Len())
		}
		probSum += p
	}

	// Could happen due to rounding errors.
	return e.Subset(0, e.Len()-1), e.Subset(e.Len()-1, e.Len())
}

func (e *event) probSum() float64 {
	var res float64
	for _, p := range e.probs {
		res += p
	}
	return res
}

// arrangePoles rearranges the bytes so that the most
// likely byte comes first and the second most likely
// byte comes last.
func (e *event) arrangePoles() {
	var biggest float64
	var biggestIdx int
	for i, x := range e.probs {
		if x > biggest {
			biggest = x
			biggestIdx = i
		}
	}
	e.Swap(biggestIdx, 0)

	biggest = 0
	biggestIdx = 1
	for i, x := range e.probs[1:] {
		if x > biggest {
			biggest = x
			biggestIdx = i + 1
		}
	}
	e.Swap(biggestIdx, e.Len()-1)
}
