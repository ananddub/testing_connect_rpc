import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TodoService, ActionType, EventType, type TodoItem } from "../gen/todo/v1/todo_pb";

export { TodoService, ActionType, EventType };
export type { TodoItem };

// ConnectRPC client over HTTPS / HTTP/2 (ALPN h2)
const transport = createConnectTransport({
  baseUrl: "https://localhost:8085",
});

export const todoClient = createClient(TodoService, transport);
