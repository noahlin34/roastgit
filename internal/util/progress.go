package util

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Spinner provides a simple progress indicator.
type Spinner struct {
	mu      sync.Mutex
	writer  io.Writer
	message string
	ticker  *time.Ticker
	done    chan struct{}
	active  bool
}

func NewSpinner(w io.Writer, message string) *Spinner {
	return &Spinner{writer: w, message: message}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		return
	}
	s.active = true
	s.done = make(chan struct{})
	s.ticker = time.NewTicker(120 * time.Millisecond)
	go func() {
		frames := []rune{'|', '/', '-', '\\'}
		i := 0
		for {
			select {
			case <-s.ticker.C:
				fmt.Fprintf(s.writer, "\r%s %c", s.message, frames[i%len(frames)])
				i++
			case <-s.done:
				return
			}
		}
	}()
}

func (s *Spinner) Stop(finalMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}
	s.active = false
	s.ticker.Stop()
	close(s.done)
	if finalMsg != "" {
		fmt.Fprintf(s.writer, "\r%s\n", finalMsg)
	} else {
		fmt.Fprint(s.writer, "\r\n")
	}
}
