package queue

import (
	"errors"
	"sync"
)

// SizeOfQueue задает размер буферизированной очереди
const SizeOfQueue = 300

// Queue интерфейс очереди
type Queue interface {
	Push(link string) error
	Pop() (string, bool)
}

// impl реализация интерфейса очереди
type impl struct {
	links map[string]bool // Используется как множество ссылок -- только уникальные
	queue chan string     // Очередь в виде канала строк
	mu    sync.Mutex      // Синхронизация для конкурентного использование map
}

// NewQueue инициализирует реализацию интерфейса Queue
func NewQueue() Queue {
	return &impl{
		links: make(map[string]bool),
		queue: make(chan string, SizeOfQueue),
	}
}

// Push помещает элемент в конец очереди
func (q *impl) Push(link string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.links[link]; ok {
		return errors.New("link already exists")
	}

	q.links[link] = true
	q.queue <- link
	return nil
}

// Pop возвращает первый элемент в очереди, если он есть
func (q *impl) Pop() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return "", false
	}

	return <-q.queue, true
}
