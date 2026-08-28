import { createClient } from "@connectrpc/connect";
import { createConnectWebSocketTransport } from "@sudorandom/connect-bidi-web";
import { TodoService, ActionType, EventType, type TodoItem } from "../gen/todo/v1/todo_pb";

export { ActionType, EventType };
export type { TodoItem };

const transport = createConnectWebSocketTransport({ baseUrl: "ws://localhost:8085" });
export const todoClient = createClient(TodoService, transport);
