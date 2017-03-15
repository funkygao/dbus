package engine

import (
	"sync"
	"testing"
)

func BenchmarkGolangChannelBuffer100(b *testing.B) {
	c := make(chan struct{}, 100)
	go func() {
		for i := 0; i < b.N; i++ {
			c <- struct{}{}
		}
	}()

	for i := 0; i < b.N; i++ {
		<-c
	}
}

func BenchmarkGolangChannelBuffer1000(b *testing.B) {
	c := make(chan struct{}, 1000)
	go func() {
		for i := 0; i < b.N; i++ {
			c <- struct{}{}
		}
	}()

	for i := 0; i < b.N; i++ {
		<-c
	}
}

func BenchmarkGolangChannelBuffer10000(b *testing.B) {
	c := make(chan struct{}, 10000)
	go func() {
		for i := 0; i < b.N; i++ {
			c <- struct{}{}
		}
	}()

	for i := 0; i < b.N; i++ {
		<-c
	}
}

func BenchmarkChannel(b *testing.B) {
	ch := NewChannel()
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			ch.Get()
		}
	}()

	for i := 0; i < b.N; i++ {
		ch.Put('X')
	}
}

func BenchmarkChannelContention(b *testing.B) {
	ch := NewChannel()
	var wg sync.WaitGroup
	var concurrency = 100
	wg.Add(concurrency)
	b.ResetTimer()

	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				ch.Put('X')
			}
		}()
	}

	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				ch.Get()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkGolangChannelUnbuffered(b *testing.B) {
	ch := make(chan interface{})
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			<-ch
		}
	}()

	for i := 0; i < b.N; i++ {
		ch <- 'X'
	}
}

func BenchmarkGolangChannelReadContention(b *testing.B) {
	ch := make(chan interface{}, 100)
	var wg sync.WaitGroup
	var concurrency = 100
	wg.Add(concurrency)
	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			ch <- 'X'
		}
	}()

	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N/concurrency; i++ {
				<-ch
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkGolangChannelContention(b *testing.B) {
	ch := make(chan interface{}, 100)
	var wg sync.WaitGroup
	var concurrency = 100
	wg.Add(concurrency)
	b.ResetTimer()

	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				ch <- 'X'
			}
		}()
	}

	for i := 0; i < concurrency; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				<-ch
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
