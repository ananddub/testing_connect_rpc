package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"
	todov1 "server/gen/todo/v1"
	"server/gen/todo/v1/todov1connect"
)

func main() {
	// Pure HTTP/3 (QUIC / UDP) Client Transport
	h3Client := &http.Client{
		Transport: &http3.RoundTripper{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := todov1connect.NewTodoServiceClient(h3Client, "https://localhost:8085")

	log.Println("⚡ [HTTP/3 QUIC Client] Sending Add Todo RPC over UDP...")
	res, err := client.ExecuteAction(context.Background(), connect.NewRequest(&todov1.TodoStreamRequest{
		Action: todov1.ActionType_ACTION_TYPE_ADD,
		Title:  "Created via Pure HTTP/3 (QUIC / UDP)!",
	}))
	if err != nil {
		log.Fatalf("❌ HTTP/3 RPC Failed: %v", err)
	}

	log.Printf("🎉 [HTTP/3 QUIC Success] Item Created: %s (ID: %s, Event: %v)",
		res.Msg.Item.Title, res.Msg.Item.Id, res.Msg.Event)
}
