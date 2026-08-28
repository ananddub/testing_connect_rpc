# ConnectRPC Real-Time Todo

Clean & modular Realtime Todo application with **ConnectRPC (gRPC)**, **Go (`gochan/broadcast`)**, and **React (TypeScript + Bun)**.

---

## 📁 Split File Structure

### 🔹 Backend (Go)
```
server/
├── main.go       # HTTP / h2c server entrypoint & CORS setup
├── service.go    # ConnectRPC TodoService implementation (ExecuteAction, SubscribeTodos)
├── store.go      # In-memory Todo store & gochan broadcast pub/sub
└── gen/          # Protobuf & ConnectRPC generated stubs
```

### 🔹 Frontend (Vite + React + Bun)
```
client/src/
├── App.tsx                    # Main container
├── components/
│   ├── TodoForm.tsx           # Add todo input component
│   ├── TodoList.tsx           # Todo list container
│   └── TodoItem.tsx           # Single todo row with toggle & delete
├── hooks/
│   └── useTodos.ts            # Custom hook for realtime stream & actions
├── lib/
│   └── connectClient.ts       # Connect client initialization
└── gen/                       # Generated Protobuf & Connect stubs
```

---

## 🚀 How to Run

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

Open [http://localhost:5173](http://localhost:5173) in multiple tabs to test real-time instant synchronization!
