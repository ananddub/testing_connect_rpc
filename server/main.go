package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/cors"
	"golang.org/x/net/http2"

	"server/gen/todo/v1/todov1connect"
)

// generateCert generates self-signed TLS cert with ALPN h2 support
func generateCert() (tls.Certificate, error) {
	if _, err1 := os.Stat("cert.pem"); err1 == nil {
		if _, err2 := os.Stat("key.pem"); err2 == nil {
			return tls.LoadX509KeyPair("cert.pem", "key.pem")
		}
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"ConnectRPC Local Dev"},
			CommonName:   "localhost",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	_ = os.WriteFile("cert.pem", certPEM, 0644)
	_ = os.WriteFile("key.pem", keyPEM, 0600)

	return tls.X509KeyPair(certPEM, keyPEM)
}

// loggingMiddleware logs incoming request protocols (HTTP/2, ALPN, Connect, gRPC, gRPC-Web)
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ignore health check from cluttering logs
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		// Detect RPC Protocol
		rpcProtocol := "HTTP / REST"
		ct := r.Header.Get("Content-Type")
		connectVer := r.Header.Get("Connect-Protocol-Version")

		if connectVer != "" {
			rpcProtocol = "Connect Protocol (v" + connectVer + ")"
		} else if strings.HasPrefix(ct, "application/grpc-web") {
			rpcProtocol = "gRPC-Web"
		} else if strings.HasPrefix(ct, "application/grpc") {
			rpcProtocol = "gRPC"
		}

		// HTTP transport protocol
		httpVersion := r.Proto // "HTTP/2.0" or "HTTP/1.1"
		alpn := "none"
		if r.TLS != nil && r.TLS.NegotiatedProtocol != "" {
			alpn = r.TLS.NegotiatedProtocol
		}

		log.Printf("🌐 [REQUEST] %s %s | Transport: %s (ALPN: %s) | RPC: %s | Content-Type: %s | Client: %s",
			r.Method,
			r.URL.Path,
			httpVersion,
			alpn,
			rpcProtocol,
			ct,
			r.RemoteAddr,
		)

		next.ServeHTTP(w, r)
	})
}

func main() {
	store := NewStore()
	service := NewTodoService(store)

	mux := http.NewServeMux()
	path, handler := todov1connect.NewTodoServiceHandler(service)
	mux.Handle(path, handler)

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK (Protocol: " + r.Proto + ")"))
	})

	// Configure CORS for Connect-Web
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

	// Wrap with protocol logging middleware
	loggedHandler := loggingMiddleware(corsHandler)

	// Generate / Load TLS Certificate
	tlsCert, err := generateCert()
	if err != nil {
		log.Fatalf("Failed to generate TLS cert: %v", err)
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
	log.Printf("🔒 ALPN: h2, http/1.1 enabled")

	if err := srv.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
