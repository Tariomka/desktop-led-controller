package common

type RingArray[Type any] struct {
	data     []Type
	capacity int
	length   int
}

func NewRingArray[Type any](capacity int) *RingArray[Type] {
	return &RingArray[Type]{
		data:     make([]Type, 0),
		capacity: capacity,
	}
}

func (ra *RingArray[Type]) Add(element Type) {
	ra.data = append([]Type{element}, ra.data...)
	if ra.length >= ra.capacity {
		ra.data = ra.data[:ra.capacity]
		return
	}
	ra.length++
}

func (ra *RingArray[Type]) Get(index int) Type { return ra.data[index] }

func (ra *RingArray[Type]) Length() int { return ra.length }
