package hashmap

import (
	"github.com/skybi/data-server/internal/task"
	"time"
)

type expiringEntry[T any] struct {
	raw      T
	inserted time.Time
}

// ExpiringMap implements the Map interface and wraps the standard NormalMap in order to implement value expiration
type ExpiringMap[K comparable, V any] struct {
	normal      *NormalMap[K, *expiringEntry[V]]
	lifetime    time.Duration
	cleanupTask *task.RepeatingTask
}

var _ Map[int, any] = (*ExpiringMap[int, any])(nil)

// NewExpiring creates a new expiring map whose values exist for a specific lifetime.
// Expired values will not be removed before ScheduleCleanupTask is called.
// Until then this map behaves exactly like a NormalMap.
func NewExpiring[K comparable, V any](lifetime time.Duration) *ExpiringMap[K, V] {
	return &ExpiringMap[K, V]{
		normal:   NewNormal[K, *expiringEntry[V]](),
		lifetime: lifetime,
	}
}

// ScheduleCleanupTask schedules the task that cleans up expired values in a specific interval.
// A call to StopCleanupTask as soon as the map is no longer needed is highly recommended because it would not be
// garbage collected otherwise.
func (obj *ExpiringMap[K, V]) ScheduleCleanupTask(tick time.Duration) {
	if obj.cleanupTask != nil {
		return
	}
	obj.cleanupTask = task.NewRepeating(func() {
		obj.normal.BootstrappedManipulation(func(raw map[K]*expiringEntry[V]) {
			for key, val := range raw {
				if time.Since(val.inserted) > obj.lifetime {
					delete(raw, key)
				}
			}
		})
	}, tick)
	obj.cleanupTask.Start()
}

// StopCleanupTask stops the cleanup task
func (obj *ExpiringMap[K, V]) StopCleanupTask() {
	obj.cleanupTask.Stop(true)
	obj.cleanupTask = nil
}

// Size returns the amount of stored key-value pairs
func (obj *ExpiringMap[K, V]) Size() int {
	return obj.normal.Size()
}

// Has returns whether a value is assigned to the given key
func (obj *ExpiringMap[K, V]) Has(key K) bool {
	_, ok := obj.Lookup(key)
	return ok
}

// Lookup returns the value assigned to the given key and a boolean indicating if the value was set manually or is
// the type's zero value
func (obj *ExpiringMap[K, V]) Lookup(key K) (V, bool) {
	val, ok := obj.normal.Lookup(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.raw, true
}

// Get returns the value assigned to the given key.
// Will be nil if it was not set using Set before.
func (obj *ExpiringMap[K, V]) Get(key K) V {
	raw := obj.normal.Get(key)
	if raw == nil {
		var zero V
		return zero
	}
	return raw.raw
}

// Set sets a key-value pair
func (obj *ExpiringMap[K, V]) Set(key K, value V) {
	obj.normal.Set(key, &expiringEntry[V]{
		raw:      value,
		inserted: time.Now(),
	})
}

// Unset deletes the value assigned to given key
func (obj *ExpiringMap[K, V]) Unset(key K) {
	obj.normal.Unset(key)
}

// Clear clears the whole map (essentially re-creating the underlying map)
func (obj *ExpiringMap[K, V]) Clear() {
	obj.normal.Clear()
}

// BootstrappedManipulation allows a thread safe direct manipulation of the underlying map by wrapping the given
// function in a lock of the underlying mutex.
// In this special case, this method is very expensive (the raw map has to be completely transformed 2 times).
// Avoid using this method whenever possible for expiring maps.
func (obj *ExpiringMap[K, V]) BootstrappedManipulation(action func(underlying map[K]V)) {
	obj.normal.BootstrappedManipulation(func(raw map[K]*expiringEntry[V]) {
		transformed := make(map[K]V, len(raw))
		for key, val := range raw {
			transformed[key] = val.raw
		}
		action(transformed)
		for key, val := range transformed {
			raw[key] = &expiringEntry[V]{
				raw:      val,
				inserted: raw[key].inserted,
			}
		}
	})
}
