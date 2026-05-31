package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"titanbot/engine"
	"titanbot/reporter"
)

func main() {
	fmt.Println("Starting TitanBot...")

	target := flag.String("target", "", "Target URL or IP:Port (e.g. http://localhost:8080 or 127.0.0.1:9000)")
	protocol := flag.String("protocol", "http", "Protocol to use: http, tcp, udp, websocket, all")
	concurrency := flag.Int("concurrency", 100, "Number of concurrent connections")
	durationSecs := flag.Int("duration", 10, "Duration of the test in seconds")
	method := flag.String("method", "GET", "HTTP Method (for http protocol)")
	payload := flag.String("payload", "TitanBot-Test-Payload", "Payload to send for POST/TCP/UDP/WS")
	header := flag.String("header", "", "Custom header to send for HTTP (e.g., 'Host: api.example.com')")
	reportFile := flag.String("report", "report.html", "Output HTML report file")

	flag.Parse()

	if *target == "" {
		fmt.Println("Error: -target is required")
		flag.Usage()
		os.Exit(1)
	}

	duration := time.Duration(*durationSecs) * time.Second

	fmt.Printf("Configuration:\n")
	fmt.Printf("  Target: %s\n", *target)
	fmt.Printf("  Protocol: %s\n", *protocol)
	fmt.Printf("  Concurrency: %d\n", *concurrency)
	fmt.Printf("  Duration: %s\n", duration)

	var metrics *engine.Metrics

	fmt.Println("Initializing attack vectors...")

	switch *protocol {
	case "http":
		runner := engine.NewHTTPRunner(*concurrency, duration, *target, *method, *header)
		metrics = runner.Run()
	case "tcp":
		runner := engine.NewTCPRunner(*concurrency, duration, *target, *payload)
		metrics = runner.Run()
	case "udp":
		runner := engine.NewUDPRunner(*concurrency, duration, *target, *payload)
		metrics = runner.Run()
	case "websocket":
		runner := engine.NewWSRunner(*concurrency, duration, *target, *payload)
		metrics = runner.Run()
	case "all":
		fmt.Println("Omni Mode Engaged: Spawning all vectors simultaneously...")
		
		u, err := url.Parse(*target)
		if err != nil {
			fmt.Printf("Error parsing target URL for omni attack: %v\n", err)
			os.Exit(1)
		}
		
		hostPort := u.Host
		if hostPort == "" {
			hostPort = *target // Fallback if not a valid URL
		}

		wsScheme := "ws"
		if u.Scheme == "https" {
			wsScheme = "wss"
		}
		wsTarget := fmt.Sprintf("%s://%s%s", wsScheme, hostPort, u.Path)

		metrics = engine.NewMetrics()
		var wg sync.WaitGroup
		results := make(chan *engine.Metrics, 4)

		wg.Add(4)
		go func() {
			defer wg.Done()
			results <- engine.NewHTTPRunner(*concurrency, duration, *target, *method, *header).Run()
		}()
		go func() {
			defer wg.Done()
			results <- engine.NewTCPRunner(*concurrency, duration, hostPort, *payload).Run()
		}()
		go func() {
			defer wg.Done()
			results <- engine.NewUDPRunner(*concurrency, duration, hostPort, *payload).Run()
		}()
		go func() {
			defer wg.Done()
			results <- engine.NewWSRunner(*concurrency, duration, wsTarget, *payload).Run()
		}()

		wg.Wait()
		close(results)

		for res := range results {
			metrics.Merge(res)
		}

	default:
		fmt.Printf("Error: Unsupported protocol '%s'\n", *protocol)
		os.Exit(1)
	}

	fmt.Println("Test completed. Generating report...")

	err := reporter.GenerateHTMLReport(metrics, *reportFile)
	if err != nil {
		fmt.Printf("Failed to generate report: %v\n", err)
	}
}
