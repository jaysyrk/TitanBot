package engine

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPRunner struct {
	*Runner
	Method string
	Header string
}

func NewHTTPRunner(concurrency int, duration time.Duration, target string, method string, header string) *HTTPRunner {
	return &HTTPRunner{
		Runner: NewRunner(concurrency, duration, target),
		Method: method,
		Header: header,
	}
}

func (r *HTTPRunner) Run() *Metrics {
	var wg sync.WaitGroup
	var collectorWg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), r.Duration)
	defer cancel()

	collectorWg.Add(1)
	go r.StartCollector(&collectorWg)

	transport := &http.Transport{
		MaxIdleConns:        r.Concurrency,
		MaxIdleConnsPerHost: r.Concurrency,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Parse header if provided "Key: Value"
	var headerKey, headerVal string
	if r.Header != "" {
		parts := strings.SplitN(r.Header, ":", 2)
		if len(parts) == 2 {
			headerKey = strings.TrimSpace(parts[0])
			headerVal = strings.TrimSpace(parts[1])
		} else {
			// If no colon, assume it's the Host header
			headerKey = "Host"
			headerVal = strings.TrimSpace(r.Header)
		}
	}

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
					req, err := http.NewRequestWithContext(ctx, r.Method, r.Target, nil)
					if err != nil {
						r.Results <- Result{Success: false, Latency: time.Since(start)}
						continue
					}
					
					if headerKey != "" {
						if strings.EqualFold(headerKey, "Host") {
							req.Host = headerVal
						} else {
							req.Header.Set(headerKey, headerVal)
						}
					}

					resp, err := client.Do(req)
					latency := time.Since(start)
					if err != nil {
						r.Results <- Result{Success: false, Latency: latency}
					} else {
						io.Copy(io.Discard, resp.Body)
						resp.Body.Close()
						success := resp.StatusCode >= 200 && resp.StatusCode < 400
						r.Results <- Result{Success: success, Latency: latency}
					}
				}
			}
		}()
	}

	wg.Wait()
	close(r.Results)
	collectorWg.Wait()

	return r.Metrics
}
