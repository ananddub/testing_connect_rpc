import React from "react";
import type { TodoItem as TodoItemType } from "../lib/connectClient";

interface TodoItemProps {
  todo: TodoItemType;
  onToggle: (todo: TodoItemType) => void;
  onDelete: (id: string) => void;
}

export function TodoItem({ todo, onToggle, onDelete }: TodoItemProps) {
  return (
    <li className="todo-item">
      <label>
        <input
          type="checkbox"
          checked={todo.completed}
          onChange={() => onToggle(todo)}
        />
        <span className={`title ${todo.completed ? "done" : ""}`}>
          {todo.title}
        </span>
      </label>

      <button className="delete-btn" onClick={() => onDelete(todo.id)} title="Delete">
        ✕
      </button>
    </li>
  );
}
