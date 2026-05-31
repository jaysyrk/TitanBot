package reporter

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"titanbot/engine"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>TitanBot Load Test Report</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0f172a;
            --card-bg: #1e293b;
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --accent: #3b82f6;
            --success: #10b981;
            --danger: #ef4444;
        }
        body {
            font-family: 'Inter', sans-serif;
            background-color: var(--bg-color);
            color: var(--text-main);
            margin: 0;
            padding: 40px;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
        }
        header {
            text-align: center;
            margin-bottom: 40px;
        }
        h1 {
            font-size: 3rem;
            margin: 0;
            background: linear-gradient(90deg, #60a5fa, #c084fc);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        p.subtitle {
            color: var(--text-muted);
            font-size: 1.2rem;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .card {
            background-color: var(--card-bg);
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            transition: transform 0.2s;
        }
        .card:hover {
            transform: translateY(-5px);
        }
        .card h3 {
            margin: 0 0 10px 0;
            color: var(--text-muted);
            font-size: 1rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        .card .value {
            font-size: 2.5rem;
            font-weight: 800;
            margin: 0;
        }
        .value.success { color: var(--success); }
        .value.danger { color: var(--danger); }
        .value.accent { color: var(--accent); }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>TitanBot Report</h1>
            <p class="subtitle">Strength Test Results</p>
        </header>

        <div class="grid">
            <div class="card">
                <h3>Total Requests</h3>
                <p class="value accent">{{.TotalRequests}}</p>
            </div>
            <div class="card">
                <h3>Duration</h3>
                <p class="value">{{.Duration}}s</p>
            </div>
            <div class="card">
                <h3>Requests / Sec</h3>
                <p class="value accent">{{.RPS}}</p>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>Success Rate</h3>
                <p class="value success">{{.SuccessRate}}%</p>
                <p style="color: var(--text-muted); margin-top: 10px;">{{.SuccessCount}} successful</p>
            </div>
            <div class="card">
                <h3>Failure Rate</h3>
                <p class="value danger">{{.FailureRate}}%</p>
                <p style="color: var(--text-muted); margin-top: 10px;">{{.FailureCount}} failed</p>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>Min Latency</h3>
                <p class="value">{{.MinLatency}}ms</p>
            </div>
            <div class="card">
                <h3>Max Latency</h3>
                <p class="value">{{.MaxLatency}}ms</p>
            </div>
            <div class="card">
                <h3>Avg Latency</h3>
                <p class="value">{{.AvgLatency}}ms</p>
            </div>
        </div>
    </div>
</body>
</html>`

type ReportData struct {
	TotalRequests int
	Duration      float64
	RPS           float64
	SuccessRate   float64
	FailureRate   float64
	SuccessCount  int
	FailureCount  int
	MinLatency    float64
	MaxLatency    float64
	AvgLatency    float64
}

func GenerateHTMLReport(metrics *engine.Metrics, filename string) error {
	duration := metrics.EndTime.Sub(metrics.StartTime).Seconds()
	
	rps := 0.0
	if duration > 0 {
		rps = float64(metrics.TotalRequests) / duration
	}

	successRate := 0.0
	failureRate := 0.0
	if metrics.TotalRequests > 0 {
		successRate = (float64(metrics.SuccessCount) / float64(metrics.TotalRequests)) * 100
		failureRate = (float64(metrics.FailureCount) / float64(metrics.TotalRequests)) * 100
	}

	var min, max, total time.Duration
	if len(metrics.Latencies) > 0 {
		min = metrics.Latencies[0]
		for _, l := range metrics.Latencies {
			if l < min {
				min = l
			}
			if l > max {
				max = l
			}
			total += l
		}
	}
	
	avg := time.Duration(0)
	if len(metrics.Latencies) > 0 {
		avg = total / time.Duration(len(metrics.Latencies))
	}

	data := ReportData{
		TotalRequests: metrics.TotalRequests,
		Duration:      duration,
		RPS:           float64(int(rps*100)) / 100,
		SuccessRate:   float64(int(successRate*100)) / 100,
		FailureRate:   float64(int(failureRate*100)) / 100,
		SuccessCount:  metrics.SuccessCount,
		FailureCount:  metrics.FailureCount,
		MinLatency:    float64(min.Milliseconds()),
		MaxLatency:    float64(max.Milliseconds()),
		AvgLatency:    float64(avg.Milliseconds()),
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		return err
	}

	fmt.Printf("Report generated successfully at %s\n", filename)
	return nil
}
