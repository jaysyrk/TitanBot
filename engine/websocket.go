package engine

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSRunner struct {
	*Runner
	Payload []byte
}

func NewWSRunner(concurrency int, duration time.Duration, target string, payload string) *WSRunner {
	return &WSRunner{
		Runner:  NewRunner(concurrency, duration, target),
		Payload: []byte(payload),
	}
}

func (r *WSRunner) Run() *Metrics {
	var wg sync.WaitGroup
	var collectorWg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), r.Duration)
	defer cancel()

	collectorWg.Add(1)
	go r.StartCollector(&collectorWg)

	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	for i := 0; i < r.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Establish connection once per goroutine
			startConn := time.Now()
			conn, _, err := dialer.DialContext(ctx, r.Target, nil)
			if err != nil {
				r.Results <- Result{Success: false, Latency: time.Since(startConn)}
				return
			}
			defer conn.Close()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					err := conn.WriteMessage(websocket.TextMessage, r.Payload)
					latency := time.Since(start)
					
					if err != nil {
						r.Results <- Result{Success: false, Latency: latency}
						return // If connection breaks, stop this goroutine
					} else {
						r.Results <- Result{Success: true, Latency: latency}
					}
					
					// Avoid saturating CPU instantly
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	wg.Wait()
	close(r.Results)
	collectorWg.Wait()

	return r.Metrics
}
