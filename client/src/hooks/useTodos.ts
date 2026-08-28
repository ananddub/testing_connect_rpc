import { useState, useEffect, useCallback, useRef } from "react";
import { create } from "@bufbuild/protobuf";
import { todoClient, ActionType, EventType, type TodoItem } from "../lib/connectClient";
import { TodoStreamRequestSchema, type TodoStreamRequest } from "../gen/todo/v1/todo_pb";

export function useTodos() {
  const [todos, setTodos] = useState<TodoItem[]>([]);
  const [status, setStatus] = useState<string>("Connecting...");

  // send() ka ref — actions is se stream ke through bhejenge
  const sendRef = useRef<((req: TodoStreamRequest) => void) | null>(null);

  // ── BiDi Stream (WebSocket) ──────────────────────────────
  useEffect(() => {
    const abort = new AbortController();

    // Request queue — AsyncIterable banata hai streamTodos ke liye
    const pending: TodoStreamRequest[] = [];
    let notify: (() => void) | null = null;
    let closed = false;

    sendRef.current = (req) => {
      console.log("📤 [Client] Enqueuing Action to Stream:", ActionType[req.action], req);
      pending.push(req);
      notify?.();
    };

    async function* requestStream() {
      while (!closed) {
        if (pending.length === 0) {
          await new Promise<void>((r) => { notify = r; });
          notify = null;
        }
        while (pending.length > 0) {
          const nextReq = pending.shift()!;
          console.log("🚀 [Client] Yielding to WebSocket:", ActionType[nextReq.action]);
          yield nextReq;
        }
      }
    }

    (async () => {
      try {
        console.log("🔌 [Client] Connecting BiDi WebSocket stream to ws://localhost:8085...");
        setStatus("Connecting...");

        const responses = todoClient.streamTodos(requestStream(), {
          signal: abort.signal,
        });

        setStatus("Connected");
        console.log("✅ [Client] BiDi Stream Active");

        for await (const res of responses) {
          console.log("📥 [Client] Received Stream Event:", EventType[res.event], res);

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
          console.log("🔌 [Client] Stream disconnected (aborted)");
        } else {
          console.error("❌ [Client] Stream Error:", err);
          setStatus("Error: " + (err?.message || String(err)));
        }
      }
    })();

    return () => {
      closed = true;
      notify?.();
      abort.abort();
      sendRef.current = null;
    };
  }, []);

  // ── Actions ──────────────────────────────────────────────

  const addTodo = useCallback((title: string) => {
    if (!title.trim()) return;
    const req = create(TodoStreamRequestSchema, {
      action: ActionType.ADD,
      title: title.trim(),
    });
    sendRef.current?.(req);
  }, []);

  const toggleTodo = useCallback((todo: TodoItem) => {
    console.log("🔄 [Client] toggleTodo called for ID:", todo.id, "current completed:", todo.completed);
    const req = create(TodoStreamRequestSchema, {
      action: ActionType.UPDATE,
      id: todo.id,
      title: todo.title,
      completed: !todo.completed,
    });
    sendRef.current?.(req);
  }, []);

  const deleteTodo = useCallback((id: string) => {
    console.log("🗑️ [Client] deleteTodo called for ID:", id);
    const req = create(TodoStreamRequestSchema, {
      action: ActionType.DELETE,
      id,
    });
    sendRef.current?.(req);
  }, []);

  return { todos, status, addTodo, toggleTodo, deleteTodo };
}
