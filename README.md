# ConnectRPC Real-Time Todo (HTTP/3 QUIC Dual-Stack Branch)

This branch (`http3`) implements **ConnectRPC over Dual-Stack HTTP/3 (QUIC / UDP) and HTTP/2 (TCP)**:
- **QUIC / UDP Transport:** Uses `github.com/quic-go/quic-go/http3` for ultra-low latency UDP streaming and 0-RTT connections.
- **HTTP/2 Fallback & Web:** Dual-stack TCP listener on the same port with automatic `Alt-Svc: h3=":8085"` upgrade advertisement.
- **Embedded Web UI:** React UI is embedded in Go (`//go:embed all:dist`) and served from `https://localhost:8085`.

---

## 🔀 Git Branches Available

* **`http3` branch (this branch):** ConnectRPC over **HTTP/3 (QUIC over UDP)** + HTTP/2 TCP Dual-Stack
* **`http2` branch:** Native ConnectRPC over **HTTP/2** (`https://localhost:8085` / `http://localhost:8085`)
* **`websocket` branch:** Full Bi-Directional Streaming over **WebSockets** (`ws://localhost:8085`)

### Switch Branches:
```bash
# Switch to HTTP/3 branch
git checkout http3

# Switch to HTTP/2 branch
git checkout http2

# Switch to WebSocket branch
git checkout websocket
```

---

## 🚀 How to Run (HTTP/3 Branch)

### 1. Run Server (Go)
```bash
cd server
go run .
```

### 2. Test HTTP/3 (QUIC / UDP) Client
```bash
cd server
go run test_client/main.go
```
Output:
```text
⚡ [HTTP/3 QUIC Client] Sending Add Todo RPC over UDP...
🎉 [HTTP/3 QUIC Success] Item Created: Created via Pure HTTP/3 (QUIC / UDP)!
```

### 3. Open in Browser
Open [https://localhost:8085](https://localhost:8085) in your browser.
