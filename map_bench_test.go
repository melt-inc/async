package async

import (
	"sync"
	"testing"
)

func BenchmarkMapSet(b *testing.B) {
	m := new(Map[int, int])
	for i := 0; i < b.N; i++ {
		m.Set(i) <- i
	}
}

func BenchmarkMapGet(b *testing.B) {
	m := new(Map[int, int])
	for i := 0; i < b.N; i++ {
		m.Set(i) <- i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i)
	}
}

func BenchmarkMapConcurrentSet(b *testing.B) {
	m := new(Map[int, int])
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			m.Set(i) <- i
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func BenchmarkMapConcurrentGet(b *testing.B) {
	m := new(Map[int, int])
	for i := 0; i < b.N; i++ {
		m.Set(i) <- i
	}
	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			m.Get(i)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func BenchmarkGoMapSet(b *testing.B) {
	m := map[int]int{}
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
}

func BenchmarkGoMapGet(b *testing.B) {
	m := map[int]int{}
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i]
	}
}

func BenchmarkGoMapConcurrentSet(b *testing.B) {
	var mu sync.Mutex
	m := map[int]int{}
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			mu.Lock()
			defer mu.Unlock()
			m[i] = i
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func BenchmarkGoMapConcurrentGet(b *testing.B) {
	m := map[int]int{}
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			_ = m[i]
			wg.Done()
		}(i)
	}
	wg.Wait()
}
