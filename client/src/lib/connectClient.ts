import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TodoService, ActionType, EventType, type TodoItem } from "../gen/todo/v1/todo_pb";

export { TodoService, ActionType, EventType };
export type { TodoItem };

// Same-origin transport (works seamlessly when embedded in Go server or standalone)
const transport = createConnectTransport({
  baseUrl: typeof window !== "undefined" ? window.location.origin : "http://localhost:8085",
});

export const todoClient = createClient(TodoService, transport);
