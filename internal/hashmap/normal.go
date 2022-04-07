package hashmap

import "sync"

// NormalMap implements the Map interface using normal hash map behaviour.
// Basically it simply wraps the builtin map type with a RWMutex mechanism in order to provide thread safety.
type NormalMap[K comparable, V any] struct {
	mtx        sync.RWMutex
	underlying map[K]V
}

var _ Map[int, any] = (*NormalMap[int, any])(nil)

// NewNormal creates a new normal thread safe Map
func NewNormal[K comparable, V any]() *NormalMap[K, V] {
	return &NormalMap[K, V]{
		underlying: make(map[K]V),
	}
}

// Size returns the amount of stored key-value pairs
func (obj *NormalMap[K, V]) Size() int {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	return len(obj.underlying)
}

// Has returns whether a value is assigned to the given key
func (obj *NormalMap[K, V]) Has(key K) bool {
	_, ok := obj.Lookup(key)
	return ok
}

// Lookup returns the value assigned to the given key and a boolean indicating if the value was set manually or is
// the type's zero value
func (obj *NormalMap[K, V]) Lookup(key K) (V, bool) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	val, ok := obj.underlying[key]
	return val, ok
}

// Get returns the value assigned to the given key.
// May be the type's zero value if it was not set using Set before; use Has or Lookup for this information.
func (obj *NormalMap[K, V]) Get(key K) V {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	return obj.underlying[key]
}

// Set sets a key-value pair
func (obj *NormalMap[K, V]) Set(key K, value V) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	obj.underlying[key] = value
}

// Unset deletes the value assigned to given key
func (obj *NormalMap[K, V]) Unset(key K) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	delete(obj.underlying, key)
}

// Clear clears the whole map (essentially re-creating the underlying map)
func (obj *NormalMap[K, V]) Clear() {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	obj.underlying = make(map[K]V)
}

// BootstrappedManipulation allows a thread safe direct manipulation of the underlying map by wrapping the given
// function in a lock of the underlying mutex
func (obj *NormalMap[K, V]) BootstrappedManipulation(action func(underlying map[K]V)) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()
	action(obj.underlying)
}
