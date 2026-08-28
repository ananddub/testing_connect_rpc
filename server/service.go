package main

import (
	"context"
	"log"

	"connectrpc.com/connect"
	todov1 "server/gen/todo/v1"
	"server/gen/todo/v1/todov1connect"
)

type TodoService struct{ store *Store }

func NewTodoService(store *Store) *TodoService { return &TodoService{store: store} }

// 1. ExecuteAction: Single Action Dispatcher (Add, Update, Delete, Sync)
func (s *TodoService) ExecuteAction(
	ctx context.Context,
	req *connect.Request[todov1.TodoStreamRequest],
) (*connect.Response[todov1.TodoStreamResponse], error) {
	msg := req.Msg
	res := &todov1.TodoStreamResponse{}

	log.Printf("📥 [HTTP/2 Action] Action: %v, Title: %s, ID: %s, Completed: %v",
		msg.Action, msg.Title, msg.Id, msg.Completed)

	switch msg.Action {
	case todov1.ActionType_ACTION_TYPE_SYNC:
		res.Event = todov1.EventType_EVENT_TYPE_SYNC
		res.Todos = s.store.GetAll()

	case todov1.ActionType_ACTION_TYPE_ADD:
		item := s.store.Add(msg.Title)
		log.Printf("✅ [HTTP/2 Action] Added: %s (%s)", item.Title, item.Id)
		res.Event = todov1.EventType_EVENT_TYPE_ADDED
		res.Item = item
		s.store.Broadcast(res)

	case todov1.ActionType_ACTION_TYPE_UPDATE:
		item := s.store.Update(msg.Id, msg.Title, msg.Completed)
		if item != nil {
			log.Printf("✅ [HTTP/2 Action] Updated: ID=%s, Completed=%v", item.Id, item.Completed)
			res.Event = todov1.EventType_EVENT_TYPE_UPDATED
			res.Item = item
			s.store.Broadcast(res)
		}

	case todov1.ActionType_ACTION_TYPE_DELETE:
		item := s.store.Delete(msg.Id)
		if item != nil {
			log.Printf("✅ [HTTP/2 Action] Deleted: ID=%s", item.Id)
			res.Event = todov1.EventType_EVENT_TYPE_DELETED
			res.Item = item
			s.store.Broadcast(res)
		}
	}

	return connect.NewResponse(res), nil
}

// 2. SubscribeTodos: Realtime Server Streaming (HTTP/2 Fetch Stream)
func (s *TodoService) SubscribeTodos(
	ctx context.Context,
	req *connect.Request[todov1.SubscribeTodosRequest],
	stream *connect.ServerStream[todov1.TodoStreamResponse],
) error {
	log.Printf("🔌 [HTTP/2 Stream] New Client subscribed: %s", req.Msg.ClientId)

	// Send initial snapshot
	initialSync := &todov1.TodoStreamResponse{
		Event: todov1.EventType_EVENT_TYPE_SYNC,
		Todos: s.store.GetAll(),
	}
	if err := stream.Send(initialSync); err != nil {
		log.Printf("❌ [HTTP/2 Stream] Failed initial sync: %v", err)
		return err
	}

	rx := s.store.Subscribe()
	defer rx.Close()

	for {
		res, err := rx.RecvContext(ctx)
		if err != nil {
			log.Printf("🔌 [HTTP/2 Stream] Subscriber disconnected: %v", err)
			return err
		}
		log.Printf("📡 [HTTP/2 Stream] Pushing event to subscriber: %v", res.Event)
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

// 3. StreamTodos: BiDi Stream (for Go/Node duplex clients)
func (s *TodoService) StreamTodos(
	ctx context.Context,
	stream *connect.BidiStream[todov1.TodoStreamRequest, todov1.TodoStreamResponse],
) error {
	return nil
}

var _ todov1connect.TodoServiceHandler = (*TodoService)(nil)
