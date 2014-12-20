package daemon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type publisher struct {
	peer
	id          int
	numMessages int
	messageSize int64
	test        test
	results     *result
	mu          sync.Mutex
}

func (p *publisher) start() {
	switch p.test {
	case throughput:
		p.testThroughput()
	case latency:
		p.testLatency()
	default:
		panic(fmt.Sprintf("Invalid test: %s", p.test))
	}
}

func (p *publisher) testThroughput() {
	message := make([]byte, p.messageSize)
	start := time.Now().UnixNano()
	for i := 0; i < p.numMessages; i++ {
		if err := p.Send(message); err != nil {
			p.mu.Lock()
			p.results = &result{Err: err.Error()}
			p.mu.Unlock()
			return
		}
	}
	stop := time.Now().UnixNano()
	ms := float32(stop-start) / 1000000
	p.mu.Lock()
	p.results = &result{
		Duration:   ms,
		Throughput: 1000 * float32(p.numMessages) / ms,
	}
	p.mu.Unlock()
	log.Println("Publisher completed")
}

func (p *publisher) testLatency() {
	message := make([]byte, 9)
	start := time.Now().UnixNano()
	for i := 0; i < p.numMessages; i++ {
		binary.PutVarint(message, time.Now().UnixNano())
		if err := p.Send(message); err != nil {
			p.mu.Lock()
			p.results = &result{Err: err.Error()}
			p.mu.Unlock()
			return
		}
	}
	stop := time.Now().UnixNano()
	ms := float32(stop-start) / 1000000
	p.mu.Lock()
	p.results = &result{
		Duration:   ms,
		Throughput: 1000 * float32(p.numMessages) / ms,
	}
	p.mu.Unlock()
	log.Println("Publisher completed")
}

func (p *publisher) getResults() (*result, error) {
	p.mu.Lock()
	r := p.results
	p.mu.Unlock()
	if r == nil {
		return nil, errors.New("Results not ready")
	}
	return r, nil
}
