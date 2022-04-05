package threadsafe

import "sync"

// Map provides a simple locked map[K]V in order to make it thread safe
type Map[K comparable, V any] struct {
	sync.RWMutex
	values map[K]V
}

// NewMap creates a new thread safe map
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		values: make(map[K]V),
	}
}

// Size returns the amount of stored K-V-pairs
func (safeMap *Map[K, V]) Size() int {
	return len(safeMap.values)
}

// Has checks if a specific key has an assigned value
func (safeMap *Map[K, V]) Has(key K) bool {
	_, ok := safeMap.Lookup(key)
	return ok
}

// Lookup looks up a specific key and returns the corresponding value and a boolean indicating if it was found
func (safeMap *Map[K, V]) Lookup(key K) (V, bool) {
	safeMap.Lock()
	defer safeMap.Unlock()
	val, ok := safeMap.values[key]
	return val, ok
}

// Get looks up a specific key and returns the corresponding value.
// This value will be the zero value for non-existing keys. Use Lookup if this information is important.
func (safeMap *Map[K, V]) Get(key K) V {
	safeMap.Lock()
	defer safeMap.Unlock()
	return safeMap.values[key]
}

// Set sets the value of a specific key
func (safeMap *Map[K, V]) Set(key K, val V) {
	safeMap.Lock()
	defer safeMap.Unlock()
	safeMap.values[key] = val
}

// Remove removes the value of a specific key
func (safeMap *Map[K, V]) Remove(key K) {
	safeMap.Lock()
	defer safeMap.Unlock()
	delete(safeMap.values, key)
}

// GetUnderlyingMap returns the underlying map.
// This method effectively bypasses the thread safety this structure implements.
// Manual calls to Lock and Unlock while manipulating this map are required to keep thread safety.
func (safeMap *Map[K, V]) GetUnderlyingMap() map[K]V {
	return safeMap.values
}

// Reset re-creates the underlying map
func (safeMap *Map[K, V]) Reset() {
	safeMap.Lock()
	defer safeMap.Unlock()
	safeMap.values = make(map[K]V)
}
