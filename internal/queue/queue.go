package queue

import (
	"mirror-wget/internal/normalizer"
	"sync"
)

// Item содержит ссылку и глубину рекурсии, на которой он был получен
type Item struct {
	URL   *normalizer.NormalizedUrl
	Depth int
}

// Queue интерфейс очереди
type Queue interface {
	Push(item Item) bool
	Pop() (Item, bool)
}

// sliceQueue реализация интерфейса очереди
type sliceQueue struct {
	mu    sync.Mutex
	queue []Item // Очередь в виде среза Queue
}

// NewQueue инициализирует реализацию интерфейса Queue
func NewQueue() Queue {
	return &sliceQueue{}
}

// Push помещает элемент в конец очереди
func (q *sliceQueue) Push(item Item) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, item)
	return true
}

// Pop возвращает первый элемент в очереди, если он есть
func (q *sliceQueue) Pop() (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue) == 0 {
		return Item{}, false
	}

	item := q.queue[0]
	q.queue = q.queue[1:]
	return item, true
}
