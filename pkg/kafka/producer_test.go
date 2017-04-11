package kafka

import (
	"testing"
)

func BenchmarkChannelSend(b *testing.B) {
	c := make(chan struct{}, 100)
	go func() {
		for range c {
		}
	}()

	for i := 0; i < b.N; i++ {
		c <- struct{}{}
	}
	close(c)
}

func BenchmarkChannelSendWithSelect(b *testing.B) {
	c := make(chan struct{}, 100)
	go func() {
		for range c {
		}
	}()

	for i := 0; i < b.N; i++ {
		select {
		case c <- struct{}{}:
		}
	}
	close(c)
}

func BenchmarkChannelSendCheckStopper(b *testing.B) {
	c := make(chan struct{}, 100)
	stopper := make(chan struct{})
	go func() {
		for range c {
		}
	}()

	for i := 0; i < b.N; i++ {
		select {
		case <-stopper:
		case c <- struct{}{}:
		}
	}
	close(c)
	close(stopper)
}

func BenchmarkSelectChannelDefault(b *testing.B) {
	c := make(chan struct{}, 100)
	stopper := make(chan struct{})
	go func() {
		for range c {
		}
	}()

	for i := 0; i < b.N; i++ {
		select {
		case <-stopper:
		default:
		}

		c <- struct{}{}
	}
	close(c)
	close(stopper)
}
