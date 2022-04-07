package hashmap

// Map represents the interface every map provided by this package has to implement
type Map[K comparable, V any] interface {
	// Size returns the amount of stored key-value pairs
	Size() int

	// Has returns whether a value is assigned to the given key
	Has(key K) bool

	// Lookup returns the value assigned to the given key and a boolean indicating if the value was set manually or is
	// the type's zero value
	Lookup(key K) (V, bool)

	// Get returns the value assigned to the given key.
	// May be the type's zero value if it was not set using Set before; use Has or Lookup for this information.
	Get(key K) V

	// Set sets a key-value pair
	Set(key K, value V)

	// Unset deletes the value assigned to given key
	Unset(key K)

	// Clear clears the whole map (essentially re-creating the underlying map)
	Clear()

	// BootstrappedManipulation allows a thread safe direct manipulation of the underlying map by wrapping the given
	// function in a lock of the underlying mutex
	BootstrappedManipulation(func(underlying map[K]V))
}
