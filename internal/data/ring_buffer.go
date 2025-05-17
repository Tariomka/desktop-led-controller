package data

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

func (this *RingBuffer[Type]) Add(element Type) {
	this.data = append([]Type{element}, this.data...)
	if this.length >= this.capacity {
		this.data = this.data[:this.capacity]
		return
	}
	this.length++
}

func (this *RingBuffer[Type]) Get(index int) Type { return this.data[index] }

func (this *RingBuffer[Type]) Length() int { return this.length }
