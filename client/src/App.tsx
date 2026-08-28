import React from "react";
import { useTodos } from "./hooks/useTodos";
import { TodoForm } from "./components/TodoForm";
import { TodoList } from "./components/TodoList";
import "./App.css";

export default function App() {
  const { todos, status, addTodo, toggleTodo, deleteTodo } = useTodos();

  const isConnected = status === "Connected";

  return (
    <div className="app">
      <div className="app-header">
        <h1>📝 Realtime Todo</h1>
        <p>BiDi Stream (ConnectRPC) — multiple tabs mein sync hota hai</p>
        <div style={{ marginTop: 8 }}>
          <span
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 6,
              fontSize: 12,
              padding: "4px 10px",
              borderRadius: 12,
              backgroundColor: isConnected ? "#dcfce7" : "#fee2e2",
              color: isConnected ? "#166534" : "#991b1b",
              fontWeight: 500,
            }}
          >
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                backgroundColor: isConnected ? "#22c55e" : "#ef4444",
              }}
            />
            {status}
          </span>
        </div>
      </div>

      <div className="card">
        <TodoForm onAdd={addTodo} />
        <TodoList todos={todos} onToggle={toggleTodo} onDelete={deleteTodo} />
      </div>
    </div>
  );
}
