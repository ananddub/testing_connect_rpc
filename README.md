# ConnectRPC Real-Time Todo (HTTP/2 Default Branch)

This branch (`http2`) implements the **default ConnectRPC architecture over HTTP/2 (Cleartext / h2c)**:
- **Server Streaming:** Browser subscribes to real-time updates via standard Fetch `ReadableStream` (`SubscribeTodos`).
- **Action Dispatcher:** Browser executes Add, Toggle, and Delete actions via standard Connect Unary RPCs (`ExecuteAction`).
- **Pub/Sub:** Backend uses `github.com/amorey/gochan/broadcast` for instant thread-safe broadcast to all connected streams.

---

## 🔀 Git Branches Available

* **`websocket` branch:** Full Bi-Directional Streaming over WebSockets (`ws://localhost:8085`)
* **`http2` branch (this branch):** Native ConnectRPC over HTTP/2 (`http://localhost:8085`)

### Switch Branches:
```bash
# Switch to WebSocket branch
git checkout websocket

# Switch to HTTP/2 branch
git checkout http2
```

---

## 🚀 How to Run (HTTP/2 Branch)

### 1. Run Backend (Go)
```bash
cd server
go run .
```

### 2. Run Frontend (Bun)
```bash
cd client
bun run dev
```

Open [http://localhost:5173](http://localhost:5173) in your browser.
Open multiple tabs to watch real-time synchronisation across clients!
