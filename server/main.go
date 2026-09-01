package main

import (
	"crypto/tls"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"server/gen/todo/v1/todov1connect"
)

//go:embed all:dist
var distFS embed.FS

// loggingMiddleware logs incoming request protocols
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ignore health check from cluttering logs
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Detect RPC Protocol
		rpcProtocol := "HTTP / Static Asset"
		ct := r.Header.Get("Content-Type")
		connectVer := r.Header.Get("Connect-Protocol-Version")

		if connectVer != "" {
			rpcProtocol = "Connect Protocol (v" + connectVer + ")"
		} else if strings.HasPrefix(ct, "application/grpc-web") {
			rpcProtocol = "gRPC-Web"
		} else if strings.HasPrefix(ct, "application/grpc") {
			rpcProtocol = "gRPC"
		}

		alpn := "none"
		if r.TLS != nil && r.TLS.NegotiatedProtocol != "" {
			alpn = r.TLS.NegotiatedProtocol
		}

		log.Printf("🌐 [REQUEST] %s %s | Transport: %s (ALPN: %s) | RPC: %s | Client: %s",
			r.Method,
			r.URL.Path,
			r.Proto,
			alpn,
			rpcProtocol,
			r.RemoteAddr,
		)

		next.ServeHTTP(w, r)
	})
}

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
		_, _ = w.Write([]byte("OK (Protocol: " + r.Proto + ")"))
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

	// Protocol logging middleware
	loggedHandler := loggingMiddleware(corsHandler)

	// Load TLS Certificates
	tlsCert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("Failed to load TLS cert: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	// TLS Config with explicit "h2" ALPN
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h2", "http/1.1"},
	}

	srv := &http.Server{
		Addr:      ":" + port,
		Handler:   loggedHandler,
		TLSConfig: tlsConfig,
	}

	// Configure HTTP/2 server
	if err := http2.ConfigureServer(srv, &http2.Server{}); err != nil {
		log.Fatalf("Failed to configure HTTP/2: %v", err)
	}

	log.Printf("🚀 ConnectRPC HTTPS / HTTP/2 Server: https://localhost:%s", port)
	log.Printf("📡 Service endpoint: %s", path)
	log.Printf("💻 Embedded Web UI: https://localhost:%s", port)
	log.Printf("🔒 ALPN: ['h2', 'http/1.1'] enabled")

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
