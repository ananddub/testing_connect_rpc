import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TodoService, ActionType, EventType, type TodoItem } from "../gen/todo/v1/todo_pb";

export { TodoService, ActionType, EventType };
export type { TodoItem };

// Original standard ConnectRPC transport over HTTP
const transport = createConnectTransport({
  baseUrl: "http://localhost:8085",
});

export const todoClient = createClient(TodoService, transport);
