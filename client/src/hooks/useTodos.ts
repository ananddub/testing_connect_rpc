import { useState, useEffect, useCallback } from "react";
import { create } from "@bufbuild/protobuf";
import { todoClient, ActionType, EventType, type TodoItem } from "../lib/connectClient";
import {
  TodoStreamRequestSchema,
  SubscribeTodosRequestSchema,
} from "../gen/todo/v1/todo_pb";

export function useTodos() {
  const [todos, setTodos] = useState<TodoItem[]>([]);
  const [status, setStatus] = useState<string>("Connecting...");

  // ── 1. Realtime Server Stream (HTTP/2 Fetch ReadableStream) ──
  useEffect(() => {
    const abort = new AbortController();

    (async () => {
      try {
        console.log("🔌 [HTTP/2 Client] Subscribing to realtime stream...");
        setStatus("Connecting...");

        const stream = todoClient.subscribeTodos(
          create(SubscribeTodosRequestSchema, { clientId: "web-client" }),
          { signal: abort.signal }
        );

        setStatus("Connected (HTTP/2)");
        console.log("✅ [HTTP/2 Client] Realtime Stream Active");

        for await (const res of stream) {
          console.log("📥 [HTTP/2 Client] Stream Event:", EventType[res.event], res);

          if (res.event === EventType.SYNC) {
            setTodos(res.todos);
          } else if (res.event === EventType.ADDED && res.item) {
            const item = res.item;
            setTodos((prev) => [...prev.filter((t) => t.id !== item.id), item]);
          } else if (res.event === EventType.UPDATED && res.item) {
            const item = res.item;
            setTodos((prev) => prev.map((t) => (t.id === item.id ? item : t)));
          } else if (res.event === EventType.DELETED && res.item) {
            const item = res.item;
            setTodos((prev) => prev.filter((t) => t.id !== item.id));
          }
        }
      } catch (err: any) {
        if (abort.signal.aborted) {
          console.log("🔌 [HTTP/2 Client] Stream disconnected");
        } else {
          console.error("❌ [HTTP/2 Client] Stream Error:", err);
          setStatus("Error: " + (err?.message || String(err)));
        }
      }
    })();

    return () => {
      abort.abort();
    };
  }, []);

  // ── 2. Action Dispatchers (HTTP/2 Unary RPCs) ────────────────

  const addTodo = useCallback(async (title: string) => {
    if (!title.trim()) return;
    try {
      const req = create(TodoStreamRequestSchema, {
        action: ActionType.ADD,
        title: title.trim(),
      });
      console.log("📤 [HTTP/2 Client] Dispatching ADD:", title);
      await todoClient.executeAction(req);
    } catch (err) {
      console.error("Failed to add todo:", err);
    }
  }, []);

  const toggleTodo = useCallback(async (todo: TodoItem) => {
    try {
      const req = create(TodoStreamRequestSchema, {
        action: ActionType.UPDATE,
        id: todo.id,
        title: todo.title,
        completed: !todo.completed,
      });
      console.log("📤 [HTTP/2 Client] Dispatching UPDATE:", todo.id, !todo.completed);
      await todoClient.executeAction(req);
    } catch (err) {
      console.error("Failed to update todo:", err);
    }
  }, []);

  const deleteTodo = useCallback(async (id: string) => {
    try {
      const req = create(TodoStreamRequestSchema, {
        action: ActionType.DELETE,
        id,
      });
      console.log("📤 [HTTP/2 Client] Dispatching DELETE:", id);
      await todoClient.executeAction(req);
    } catch (err) {
      console.error("Failed to delete todo:", err);
    }
  }, []);

  return { todos, status, addTodo, toggleTodo, deleteTodo };
}
