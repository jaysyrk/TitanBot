package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"titanbot/engine"
	"titanbot/reporter"
)

func verifyAuthorization(target string, expectedToken string) {
	u, err := url.Parse(target)
	if err != nil {
		fmt.Printf("Error parsing target for authorization: %v\n", err)
		os.Exit(1)
	}
	
	host := u.Host
	if host == "" {
		host = target
	}
	
	scheme := "http"
	if u.Scheme == "https" || u.Scheme == "wss" {
		scheme = "https"
	}
	authURL := fmt.Sprintf("%s://%s/.well-known/load-test-auth.txt", scheme, host)
	
	fmt.Printf("Verifying authorization lock at: %s\n", authURL)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	
	resp, err := client.Get(authURL)
	if err != nil {
		fmt.Printf("Error: Target not authorized for load testing. Handshake failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		fmt.Printf("Error: Target not authorized for load testing. Handshake failed with status: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: Failed to read auth token from target: %v\n", err)
		os.Exit(1)
	}
	
	body := strings.TrimSpace(string(bodyBytes))
	if body != expectedToken {
		fmt.Printf("Error: Target not authorized for load testing. Token mismatch.\nExpected: %s\nGot: %s\n", expectedToken, body)
		os.Exit(1)
	}
	
	fmt.Println("Authorization lock passed. Target verified.")
}

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
	authToken := flag.String("auth-token", "", "Authorization token required to unlock the target server for testing")

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

	if *authToken == "" {
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			fmt.Printf("Error generating random auth token: %v\n", err)
			os.Exit(1)
		}
		generatedToken := hex.EncodeToString(bytes)
		fmt.Println("--------------------------------------------------")
		fmt.Println("AUTHORIZATION LOCK ENGAGED")
		fmt.Println("To prevent unauthorized DDoS attacks, TitanBot requires a handshake.")
		fmt.Printf("Please create a file at /.well-known/load-test-auth.txt on your target server.\n")
		fmt.Printf("The file must contain exactly this token: %s\n\n", generatedToken)
		fmt.Printf("Press Enter to continue once the file is uploaded...")
		fmt.Scanln()
		fmt.Println("--------------------------------------------------")
		*authToken = generatedToken
	}

	verifyAuthorization(*target, *authToken)

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
