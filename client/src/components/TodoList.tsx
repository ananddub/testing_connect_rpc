import React from "react";
import type { TodoItem as TodoItemType } from "../lib/connectClient";
import { TodoItem } from "./TodoItem";

interface TodoListProps {
  todos: TodoItemType[];
  onToggle: (todo: TodoItemType) => void;
  onDelete: (id: string) => void;
}

export function TodoList({ todos, onToggle, onDelete }: TodoListProps) {
  if (todos.length === 0) {
    return <p className="empty">Koi todo nahi hai — upar se add karo!</p>;
  }

  return (
    <ul className="todo-list">
      {todos.map((todo) => (
        <TodoItem key={todo.id} todo={todo} onToggle={onToggle} onDelete={onDelete} />
      ))}
    </ul>
  );
}
