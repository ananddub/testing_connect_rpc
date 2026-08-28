// Code generated for connectrpc.com/connect/v2
// Source: todo/v1/todo.proto

package todov1connect

import (
	connect "connectrpc.com/connect/v2"
	context "context"
	sync "sync"
	v1 "server/gen/todo/v1"
)

const TodoServiceName = "todo.v1.TodoService"

const TodoServiceStreamTodosProcedure = "/todo.v1.TodoService/StreamTodos"

var todoServiceStreamTodosSpec = sync.OnceValue(func() connect.Spec {
	return connect.Spec{
		StreamType: connect.StreamTypeBidi,
		Schema:     v1.File_todo_v1_todo_proto.Services().ByName("TodoService").Methods().ByName("StreamTodos"),
		Procedure:  TodoServiceStreamTodosProcedure,
	}
})

// TodoServiceStreamTodosServerStream — bidi stream type
type TodoServiceStreamTodosServerStream struct{ stream connect.ServerStream }

func (s TodoServiceStreamTodosServerStream) Receive() (*v1.TodoStreamRequest, error) {
	var req v1.TodoStreamRequest
	if err := s.stream.Receive(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s TodoServiceStreamTodosServerStream) Send(res *v1.TodoStreamResponse) error {
	return s.stream.Send(res)
}

// TodoServiceHandler — is interface ko implement karo
type TodoServiceHandler interface {
	StreamTodos(context.Context, TodoServiceStreamTodosServerStream) error
}

// RegisterTodoServiceHandler — connect.Server mein register karo
func RegisterTodoServiceHandler(server *connect.Server, svc TodoServiceHandler) {
	adapter := todoServiceAdapter{svc: svc}
	server.Register(
		connect.Method{Spec: todoServiceStreamTodosSpec(), Handler: adapter.streamTodos},
	)
}

type todoServiceAdapter struct{ svc TodoServiceHandler }

func (a todoServiceAdapter) streamTodos(ctx context.Context, _ connect.Spec, stream connect.ServerStream) error {
	return a.svc.StreamTodos(ctx, TodoServiceStreamTodosServerStream{stream: stream})
}
