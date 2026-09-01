package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"server/gen/todo/v1/todov1connect"
)

//go:embed all:dist
var distFS embed.FS

func main() {
	store := NewStore()
	service := NewTodoService(store)

	mux := http.NewServeMux()

	// 1. ConnectRPC Service Route
	path, handler := todov1connect.NewTodoServiceHandler(service)
	mux.Handle(path, handler)

	// 2. Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 3. Embedded React Frontend SPA Handler
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("Failed to load embedded frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(subFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Don't intercept RPC endpoints
		if strings.HasPrefix(r.URL.Path, "/todo.v1.") {
			handler.ServeHTTP(w, r)
			return
		}

		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath != "" {
			// Check if file exists in embedded dist (e.g. assets/...)
			f, err := subFS.Open(cleanPath)
			if err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA root and client-side routing
		indexData, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexData)
	})

	// CORS Configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{
			"Connect-Protocol-Version",
			"Connect-Content-Encoding",
			"Grpc-Status",
			"Grpc-Message",
			"Grpc-Status-Details-Bin",
		},
		MaxAge: 7200,
	}).Handler(mux)

	// Standard ConnectRPC HTTP/2 Cleartext (h2c) Handler
	h2cHandler := h2c.NewHandler(corsHandler, &http2.Server{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	log.Printf("🚀 All-in-One ConnectRPC & Embedded React Server: http://localhost:%s", port)
	log.Printf("📡 Service endpoint: %s", path)
	log.Printf("💻 Embedded Web UI: http://localhost:%s", port)

	if err := http.ListenAndServe(":"+port, h2cHandler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
