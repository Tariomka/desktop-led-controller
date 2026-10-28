package data

type Tuple[K any, V any] struct {
	Key   K
	Value V
}

func NewTuple[K any, V any](key K, value V) Tuple[K, V] {
	return Tuple[K, V]{
		Key:   key,
		Value: value,
	}
}
