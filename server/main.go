package main

import (
	"log"
	"net/http"
	"os"

	connect "connectrpc.com/connect/v2"
	"github.com/coder/websocket"
	"github.com/rs/cors"
	"github.com/sudorandom/connect-bidi-web/connectwebsocket"
	"server/gen/todo/v1/todov1connect"
)

func main() {
	store := NewStore()
	service := NewTodoService(store)

	srv := connect.NewServer()
	todov1connect.RegisterTodoServiceHandler(srv, service)

	// WebSocket handler — browser mein bidi stream ke liye
	wsHandler := connectwebsocket.NewHandler(srv, connectwebsocket.WithAcceptOptions(
		&websocket.AcceptOptions{InsecureSkipVerify: true},
	))

	mux := http.NewServeMux()
	mux.Handle("/", wsHandler)

	corsHandler := cors.AllowAll().Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("🚀 Server: http://localhost:%s", port)
	log.Printf("🔌 WebSocket: ws://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
