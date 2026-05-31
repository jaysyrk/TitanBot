package engine

import (
	"context"
	"net"
	"sync"
	"time"
)

type TCPRunner struct {
	*Runner
	Payload []byte
}

func NewTCPRunner(concurrency int, duration time.Duration, target string, payload string) *TCPRunner {
	return &TCPRunner{
		Runner:  NewRunner(concurrency, duration, target),
		Payload: []byte(payload),
	}
}

func (r *TCPRunner) Run() *Metrics {
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
			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					// Create connection
					conn, err := net.DialTimeout("tcp", r.Target, 5*time.Second)
					if err != nil {
						r.Results <- Result{Success: false, Latency: time.Since(start)}
						continue
					}
					
					conn.SetDeadline(time.Now().Add(5 * time.Second))
					
					// Send payload
					_, err = conn.Write(r.Payload)
					latency := time.Since(start)
					
					if err != nil {
						r.Results <- Result{Success: false, Latency: latency}
					} else {
						// Wait for some response or just consider it successful write
						buf := make([]byte, 1024)
						conn.Read(buf) // Ignore read errors, mostly care about connection and write
						
						r.Results <- Result{Success: true, Latency: latency}
					}
					
					conn.Close()
				}
			}
		}()
	}

	wg.Wait()
	close(r.Results)
	collectorWg.Wait()

	return r.Metrics
}
