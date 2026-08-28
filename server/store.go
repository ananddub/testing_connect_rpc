package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/amorey/gochan/broadcast"
	todov1 "server/gen/todo/v1"
)

// Store — yeh poori app ka "data center" hai
// Todos memory mein store hote hain (database nahi hai)
// Aur jab bhi koi change hota hai, sabko broadcast hota hai (pub/sub)
type Store struct {
	mu      sync.RWMutex                               // Thread safety ke liye lock (multiple goroutines ek saath access kar sakti hain)
	todos   []*todov1.TodoItem                          // Todos ki list (slice)
	hub     *broadcast.Hub[*todov1.TodoStreamResponse] // Pub/sub bus — connected clients ko live updates bhejta hai
	counter int64                                      // Naye todo ka ID banane ke liye auto-increment counter
}

// NewStore — nayi store banata hai with kuch sample todos
func NewStore() *Store {
	return &Store{
		todos: []*todov1.TodoItem{
			{
				Id:        "1",
				Title:     "Learn ConnectRPC with Go & React",
				Completed: true,
				CreatedAt: time.Now().UnixMilli(),
			},
			{
				Id:        "2",
				Title:     "Realtime pub/sub with gochan",
				Completed: false,
				CreatedAt: time.Now().UnixMilli(),
			},
		},
		// broadcast.New(128) → 128 size ka buffered channel banata hai
		// matlab 128 messages queue ho sakte hain bina kisi ko block kiye
		hub:     broadcast.New[*todov1.TodoStreamResponse](128),
		counter: 3, // 1 aur 2 already use ho gaye hain upar
	}
}

// GetAll — saari todos ki copy return karta hai
// RLock use karta hai (sirf read ke liye lock, multiple readers ek saath chal sakte hain)
func (s *Store) GetAll() []*todov1.TodoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// copy karte hain taaki bahar wala code original slice ko modify na kar sake
	list := make([]*todov1.TodoItem, len(s.todos))
	copy(list, s.todos)
	return list
}

// Add — nayi todo banata hai aur list mein add karta hai
func (s *Store) Add(title string) *todov1.TodoItem {
	s.mu.Lock() // Write lock — sirf ek hi goroutine ek waqt mein write kar sakti hai
	defer s.mu.Unlock()

	item := &todov1.TodoItem{
		Id:        fmt.Sprintf("%d", s.counter), // counter ko string ID mein convert karo
		Title:     title,
		Completed: false,
		CreatedAt: time.Now().UnixMilli(), // milliseconds mein timestamp
	}
	s.counter++
	s.todos = append(s.todos, item)
	return item
}

// Update — existing todo ko update karta hai (title ya completed status)
func (s *Store) Update(id string, title string, completed bool) *todov1.TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.todos {
		if t.Id == id {
			// Title sirf tab update karo jab empty na ho
			if title != "" {
				t.Title = title
			}
			t.Completed = completed
			return t
		}
	}
	return nil // agar todo mila hi nahi
}

// Delete — todo ko list se hata deta hai aur deleted item return karta hai
func (s *Store) Delete(id string) *todov1.TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.Id == id {
			deleted := t

			// ✅ Sahi tarika: item ke pehle wale + baad wale elements ko join karo
			// s.todos[:i]   → index i se pehle ke items
			// s.todos[i+1:] → index i ke baad ke items
			// dono ko append karke deleted item ko skip karte hain
			s.todos = append(s.todos[:i], s.todos[i+1:]...)

			return deleted
		}
	}
	return nil // agar todo mila hi nahi
}

// Broadcast — sabhi connected clients ko ek event bhejta hai
// Service is function ko call karta hai jab bhi koi add/update/delete hota hai
func (s *Store) Broadcast(res *todov1.TodoStreamResponse) {
	s.hub.Sender().Send(res)
}

// Subscribe — nayi subscription return karta hai
// Har connected client ke liye alag receiver banta hai
// Jab bhi Broadcast() call hota hai, har receiver ko wo message milta hai
func (s *Store) Subscribe() *broadcast.Receiver[*todov1.TodoStreamResponse] {
	return s.hub.Receiver()
}
