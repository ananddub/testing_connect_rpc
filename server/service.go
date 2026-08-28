package main

import (
	"context"
	"log"

	todov1 "server/gen/todo/v1"
	"server/gen/todo/v1/todov1connect"
)

type TodoService struct{ store *Store }

func NewTodoService(store *Store) *TodoService { return &TodoService{store: store} }

// StreamTodos — BiDi Stream (WebSocket ke through)
// Client → actions bhejta hai | Server → live events push karta hai
func (s *TodoService) StreamTodos(
	ctx context.Context,
	stream todov1connect.TodoServiceStreamTodosServerStream,
) error {
	log.Println("🔌 [Server] New Client connected to StreamTodos")

	// Naye client ko pehle saari todos bhejo
	todos := s.store.GetAll()
	log.Printf("📡 [Server] Sending initial SYNC with %d todos", len(todos))
	if err := stream.Send(&todov1.TodoStreamResponse{
		Event: todov1.EventType_EVENT_TYPE_SYNC,
		Todos: todos,
	}); err != nil {
		log.Printf("❌ [Server] Failed to send SYNC: %v", err)
		return err
	}

	rx := s.store.Subscribe()
	defer rx.Close()

	// Goroutine 1: client se actions padhna
	type clientMsg struct {
		req *todov1.TodoStreamRequest
		err error
	}
	clientCh := make(chan clientMsg, 1)
	go func() {
		for {
			req, err := stream.Receive()
			clientCh <- clientMsg{req, err}
			if err != nil {
				return
			}
		}
	}()

	// Goroutine 2: broadcast events receive karna
	type broadcastMsg struct {
		res *todov1.TodoStreamResponse
		err error
	}
	broadcastCh := make(chan broadcastMsg, 1)
	go func() {
		for {
			res, err := rx.RecvContext(ctx)
			broadcastCh <- broadcastMsg{res, err}
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("🔌 [Server] Client disconnected (ctx.Done)")
			return ctx.Err()

		case msg := <-clientCh:
			if msg.err != nil {
				log.Printf("🔌 [Server] Client stream closed: %v", msg.err)
				return msg.err
			}
			log.Printf("📥 [Server] Received Action: %v, Title: %s, ID: %s, Completed: %v",
				msg.req.Action, msg.req.Title, msg.req.Id, msg.req.Completed)

			switch msg.req.Action {
			case todov1.ActionType_ACTION_TYPE_ADD:
				item := s.store.Add(msg.req.Title)
				log.Printf("✅ [Server] Added Todo: ID=%s, Title=%s", item.Id, item.Title)
				s.store.Broadcast(&todov1.TodoStreamResponse{Event: todov1.EventType_EVENT_TYPE_ADDED, Item: item})

			case todov1.ActionType_ACTION_TYPE_UPDATE:
				item := s.store.Update(msg.req.Id, msg.req.Title, msg.req.Completed)
				if item != nil {
					log.Printf("✅ [Server] Updated Todo: ID=%s, Completed=%v", item.Id, item.Completed)
					s.store.Broadcast(&todov1.TodoStreamResponse{Event: todov1.EventType_EVENT_TYPE_UPDATED, Item: item})
				}

			case todov1.ActionType_ACTION_TYPE_DELETE:
				item := s.store.Delete(msg.req.Id)
				if item != nil {
					log.Printf("✅ [Server] Deleted Todo: ID=%s", item.Id)
					s.store.Broadcast(&todov1.TodoStreamResponse{Event: todov1.EventType_EVENT_TYPE_DELETED, Item: item})
				}
			}

		case msg := <-broadcastCh:
			if msg.err != nil {
				return msg.err
			}
			log.Printf("📡 [Server] Pushing event to client: %v", msg.res.Event)
			if err := stream.Send(msg.res); err != nil {
				log.Printf("❌ [Server] Failed to push event: %v", err)
				return err
			}
		}
	}
}

var _ todov1connect.TodoServiceHandler = (*TodoService)(nil)
