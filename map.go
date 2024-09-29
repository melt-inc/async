package async

import "sync"

// Map can be used to get and set values in concurrent contexts. It avoids the
// common pitfalls of using sync.Map, which is not type safe and prone to
// misuse. The zero value is ready to use.
type Map[K comparable, V any] struct {
	queue  map[K][]chan<- V
	values map[K]V
	mu     sync.RWMutex
}

// NewMap initializes an async.Map with the given Go map. Since the zero value
// is ready to use, an empty map can be initialized with new(Map[K, V]).
func NewMap[K comparable, V any](m map[K]V) *Map[K, V] {
	return &Map[K, V]{
		values: m,
	}
}

// Get returns a channel that will eventually receive the value for the given
// key. If the value is already set, the channel will receive it immediately.
func (m *Map[K, V]) Get(key K) <-chan V {
	o := make(chan V, 1)
	if v, ok := m.happyPath(key); ok {
		o <- v
		close(o)
		return o
	}
	m.enqueue(key, o)
	return o
}

func (m *Map[K, V]) happyPath(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.values[key]
	return value, ok
}

// Set returns a channel that will consume a single value and store it with the
// given key. Any waiting getters for the same key will receive the value.
func (m *Map[K, V]) Set(key K) chan<- V {
	i := make(chan V, 1)
	go func() {
		m.dequeue(key, <-i)
		close(i)
	}()
	return i
}

// Delete removes the given key from the map.
func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.values, key)
}

// GetElseSet returns a channel that will receive the value from the map if it
// exists, or updates the map with the result of the given value-producing
// function. The value producer is called in a new goroutine so this method is
// non-blocking.
func (m *Map[K, V]) GetElseSet(key K, fn func() V) <-chan V {
	o := make(chan V, 1)
	if v, ok := m.happyPath(key); ok {
		o <- v
		close(o)
		return o
	}
	m.enqueue(key, o)
	go func() { m.dequeueFn(key, fn) }()
	return o
}

func (m *Map[K, V]) enqueue(key K, getter chan<- V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.queue == nil {
		m.queue = make(map[K][]chan<- V)
	}
	m.queue[key] = append(m.queue[key], getter)
}

func (m *Map[K, V]) dequeue(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// set value for future getters
	if m.values == nil {
		m.values = make(map[K]V)
	}
	m.values[key] = value

	// clear waiting getters
	for _, getter := range m.queue[key] {
		getter <- value // we know this doesn't block because the channel has a buffer of 1
		close(getter)
	}
	delete(m.queue, key)
}

func (m *Map[K, V]) dequeueFn(key K, fn func() V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	value := fn()

	// set value for future getters
	if m.values == nil {
		m.values = make(map[K]V)
	}
	m.values[key] = value

	// clear waiting getters
	for _, getter := range m.queue[key] {
		getter <- value // we know this doesn't block because the channel has a buffer of 1
		close(getter)
	}
	delete(m.queue, key)
}
