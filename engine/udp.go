package engine

import (
	"context"
	"net"
	"sync"
	"time"
)

type UDPRunner struct {
	*Runner
	Payload []byte
}

func NewUDPRunner(concurrency int, duration time.Duration, target string, payload string) *UDPRunner {
	return &UDPRunner{
		Runner:  NewRunner(concurrency, duration, target),
		Payload: []byte(payload),
	}
}

func (r *UDPRunner) Run() *Metrics {
	var wg sync.WaitGroup
	var collectorWg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), r.Duration)
	defer cancel()

	collectorWg.Add(1)
	go r.StartCollector(&collectorWg)

	for i := 0; i < r.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Resolve address once per goroutine
			addr, err := net.ResolveUDPAddr("udp", r.Target)
			if err != nil {
				return
			}
			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					
					_, err := conn.Write(r.Payload)
					latency := time.Since(start)
					
					if err != nil {
						r.Results <- Result{Success: false, Latency: latency}
					} else {
						r.Results <- Result{Success: true, Latency: latency}
					}
					
					// Small sleep to prevent eating 100% CPU on local client since UDP is so fast
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()
	close(r.Results)
	collectorWg.Wait()

	return r.Metrics
}
