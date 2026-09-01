package main

import (
	"crypto/tls"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/quic-go/quic-go/http3"
	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"server/gen/todo/v1/todov1connect"
)

//go:embed all:dist
var distFS embed.FS

// loggingMiddleware logs incoming request protocols (HTTP/3, HTTP/2, HTTP/1.1)
func loggingMiddleware(next http.Handler, port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Advertise HTTP/3 (QUIC) capability to web browsers via Alt-Svc header
		w.Header().Set("Alt-Svc", `h3=":`+port+`"; ma=86400, h3-29=":`+port+`"; ma=86400`)

		// Health check
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK (Protocol: " + r.Proto + ")"))
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

		protoName := r.Proto
		if r.ProtoMajor == 3 || strings.Contains(r.Proto, "HTTP/3") {
			protoName = "HTTP/3.0 (QUIC / UDP)"
		}

		log.Printf("🌐 [REQUEST] %s %s | Transport: %s (ALPN: %s) | RPC: %s | Client: %s",
			r.Method,
			r.URL.Path,
			protoName,
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

	// 2. Embedded React Frontend SPA Handler
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
			f, err := subFS.Open(cleanPath)
			if err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA
		indexData, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexData)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

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
			"Alt-Svc",
		},
		MaxAge: 7200,
	}).Handler(mux)

	// Protocol logging middleware with Alt-Svc
	loggedHandler := loggingMiddleware(corsHandler, port)

	// Load TLS Certificate (Required for both HTTP/2 and QUIC HTTP/3)
	tlsCert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("Failed to load TLS cert: %v", err)
	}

	// TLS Config with HTTP/3 and HTTP/2 ALPN
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3", "h3-29", "h2", "http/1.1"},
	}

	// ── 3. Start QUIC HTTP/3 Server (UDP on port 8085) ────────
	h3Server := &http3.Server{
		Addr:      ":" + port,
		Handler:   loggedHandler,
		TLSConfig: tlsConfig,
	}

	go func() {
		log.Printf("⚡ [QUIC / UDP] HTTP/3 Server running on port %s", port)
		if err := h3Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP/3 error: %v", err)
		}
	}()

	// ── 4. Start HTTPS HTTP/2 & TCP Server (TCP on port 8085) ──
	srv := &http.Server{
		Addr:      ":" + port,
		Handler:   loggedHandler,
		TLSConfig: tlsConfig,
	}
	_ = http2.ConfigureServer(srv, &http2.Server{})

	log.Printf("🚀 ConnectRPC Dual-Stack (HTTP/3 QUIC + HTTP/2 TCP): https://localhost:%s", port)
	log.Printf("📡 Service endpoint: %s", path)
	log.Printf("💻 Embedded Web UI: https://localhost:%s", port)
	log.Printf("🔒 Protocols: HTTP/3 (UDP), HTTP/2 (TCP), HTTP/1.1 (TCP)")

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("TCP Server error: %v", err)
	}
}
