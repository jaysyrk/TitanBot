# TitanBot 🚀

TitanBot is a high-performance, multi-vector load testing engine built in Go. It is designed to "strength test" enterprise web servers, gateways, and service meshes by subjecting them to massive, sustained connection floods across multiple protocols simultaneously.

## Features

- **Multi-Protocol Support:** Attack targets using HTTP/HTTPS, WebSockets, raw TCP, and UDP.
- **Omni Mode:** Fire all vectors at once (`-protocol all`) to simulate chaotic, heavy enterprise traffic from every angle.
- **High Concurrency:** Built with Go's lightweight goroutines, allowing you to spawn thousands of concurrent connections effortlessly from a single machine.
- **Customizable:** Inject custom HTTP headers (like `Host` or `Authorization`) and specify raw payloads for TCP/UDP tests.
- **Beautiful HTML Reports:** Automatically generates a sleek, dark-themed `report.html` dashboard detailing success rates, throughput (RPS), and latency distribution.

## Installation

Ensure you have Go 1.22+ installed. Clone the repository and build the binary:

```bash
git clone https://github.com/jaysyrk/TitanBot.git
cd TitanBot
go build -o titanbot.exe
```

## Usage

Run TitanBot from your terminal with the required flags.

```powershell
.\titanbot.exe -target <URL> -protocol <http|tcp|udp|websocket|all> -concurrency <number> -duration <seconds>
```

### Examples

**1. Omni Mode (Test Everything)**
Spawns 500 connections *per protocol* (2000 total) to bombard the target from all angles for 30 seconds.
```powershell
.\titanbot.exe -target https://api.example.com/v1/health -protocol all -concurrency 500 -duration 30
```

**2. Test an HTTP API with Custom Headers**
```powershell
.\titanbot.exe -target http://localhost:8080/api -protocol http -header "Host: api.internal.local" -concurrency 200 -duration 10
```

**3. Spam a WebSocket Endpoint**
```powershell
.\titanbot.exe -target wss://ws.example.com/chat -protocol websocket -payload "LoadTestMessage" -concurrency 1000 -duration 60
```

## Disclaimer

> **⚠️ WARNING:** TitanBot is an extremely aggressive load generation tool. Running high-concurrency Omni attacks against servers you do not own may be classified as a Denial of Service (DoS) attack. **Only point TitanBot at infrastructure you own or have explicit authorization to test.**
