package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"server/gen/todo/v1/todov1connect"
)

func main() {
	store := NewStore()
	service := NewTodoService(store)

	mux := http.NewServeMux()
	path, handler := todov1connect.NewTodoServiceHandler(service)
	mux.Handle(path, handler)

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Configure CORS for Connect-Web
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Content-Type",
			"Connect-Protocol-Version",
			"Connect-Timeout-Ms",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Grpc-Timeout",
			"X-Grpc-Web",
			"X-User-Agent",
		},
		ExposedHeaders: []string{
			"Connect-Protocol-Version",
			"Connect-Content-Encoding",
			"Grpc-Status",
			"Grpc-Message",
			"Grpc-Status-Details-Bin",
		},
		MaxAge: 7200,
	}).Handler(mux)

	// Default HTTP/2 Cleartext (h2c) handler
	h2cHandler := h2c.NewHandler(corsHandler, &http2.Server{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("🚀 ConnectRPC HTTP/2 Server: http://localhost:%s", port)
	log.Printf("📡 Service endpoint: %s", path)

	if err := http.ListenAndServe(":"+port, h2cHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
