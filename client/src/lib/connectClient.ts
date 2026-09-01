import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { TodoService, ActionType, EventType, type TodoItem } from "../gen/todo/v1/todo_pb";

export { TodoService, ActionType, EventType };
export type { TodoItem };

// Same-origin relative path for pure HTTPS / HTTP/2 (h2)
const transport = createConnectTransport({
  baseUrl: "",
});

export const todoClient = createClient(TodoService, transport);
