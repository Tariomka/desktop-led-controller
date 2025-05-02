package common

type RingBuffer[Type any] struct {
	data     []Type
	capacity int
	length   int
}

func NewRingBuffer[Type any](capacity int) *RingBuffer[Type] {
	return &RingBuffer[Type]{
		data:     make([]Type, 0),
		capacity: capacity,
	}
}

func (ra *RingBuffer[Type]) Add(element Type) {
	ra.data = append([]Type{element}, ra.data...)
	if ra.length >= ra.capacity {
		ra.data = ra.data[:ra.capacity]
		return
	}
	ra.length++
}

func (ra *RingBuffer[Type]) Get(index int) Type { return ra.data[index] }

func (ra *RingBuffer[Type]) Length() int { return ra.length }
