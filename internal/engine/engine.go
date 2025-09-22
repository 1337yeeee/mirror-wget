package engine

import (
	"context"
	"fmt"
	"log"
	"mirror-wget/internal/cli"
	"mirror-wget/internal/downloader"
	"mirror-wget/internal/normalizer"
	"mirror-wget/internal/queue"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const SleepDuration = 100 * time.Millisecond

// Engine структура для управления dispatcher'ом
type Engine struct {
	baseURL     *normalizer.NormalizedUrl
	queue       queue.Queue
	visited     *sync.Map
	downloadMap *sync.Map
	numWorkers  int
	maxDepth    int
	wg          *sync.WaitGroup
	activeTasks int32
	robotsTxt   *downloader.Robots
}

// NewEngine инициализирует Engine
func NewEngine(URL *normalizer.NormalizedUrl, robotsTxt *downloader.Robots, numWorkers, maxDepth int) *Engine {
	return &Engine{
		baseURL:     URL,
		queue:       queue.NewQueue(),
		visited:     &sync.Map{},
		downloadMap: &sync.Map{},
		numWorkers:  numWorkers,
		maxDepth:    maxDepth,
		wg:          &sync.WaitGroup{},
		robotsTxt:   robotsTxt,
	}
}

// Handle инициализирует и запускает Engine
func Handle() error {
	config, err := cli.NewConfig()
	if err != nil {
		return err
	}

	normURL, err := normalizer.NewNormalizedUrl(config.URL)
	if err != nil {
		return err
	}

	robotsTxt, err := downloader.LoadRobots(normURL.URL)
	if err != nil {
		return err
	}

	log.Printf("Recursion level is %d\n", config.Level)
	engine := NewEngine(normURL, robotsTxt, runtime.GOMAXPROCS(0)-1, config.Level)
	return engine.Start()
}

// Start запускает воркеры и диспатчеры
func (e *Engine) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan queue.Item, 100)
	item := queue.Item{
		URL:   e.baseURL,
		Depth: 0,
	}
	jobs <- item
	atomic.AddInt32(&e.activeTasks, 1)

	storageQueue := queue.NewQueue()

	for n := 0; n < e.numWorkers; n++ {
		e.wg.Add(1)
		w := NewWorker(e.baseURL, e.wg, &e.activeTasks, e.queue, storageQueue, e.downloadMap)
		go w.Worker(ctx, n, jobs)
	}

	e.wg.Add(1)
	go e.dispatcher(ctx, e.queue, jobs, cancel)

	log.Println("Ждем завершения worker го рутин")
	e.wg.Wait()

	jobs = make(chan queue.Item, 100)
	jobs <- item
	atomic.AddInt32(&e.activeTasks, 1)

	for n := 0; n < e.numWorkers; n++ {
		e.wg.Add(1)
		w := NewStorageWorker(e.baseURL, e.wg, storageQueue, &e.activeTasks, e.downloadMap)
		go w.Storage(ctx, n, jobs)
	}

	e.wg.Add(1)
	go e.dispatcher(ctx, storageQueue, jobs, cancel)

	log.Println("Ждем завершения storage го рутин")
	e.wg.Wait()

	e.downloadMap.Range(func(key, value interface{}) bool {
		fmt.Println(key)
		return true
	})

	return nil
}

// dispatcher управляет потоком задач для воркеров
func (e *Engine) dispatcher(ctx context.Context, itemsQueue queue.Queue, jobs chan<- queue.Item, cancel context.CancelFunc) {
	defer e.wg.Done()
	defer close(jobs)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if item, ok := itemsQueue.Pop(); ok {
				if e.maxDepth >= 0 && item.Depth > e.maxDepth {
					atomic.AddInt32(&e.activeTasks, -1)
					continue
				}
				if _, visited := e.visited.LoadOrStore(item.URL.String(), true); visited {
					atomic.AddInt32(&e.activeTasks, -1)
					continue
				}
				if e.robotsTxt != nil && !e.robotsTxt.Allowed(item.URL.URL) {
					atomic.AddInt32(&e.activeTasks, -1)
					continue
				}

				select {
				case jobs <- item:
				case <-ctx.Done():
					return
				}
			} else {
				// если очередь пуста, ждем SleepDuration
				log.Printf("Queue is empty, spleeping for: %d millisecond\n", SleepDuration.Milliseconds())
				time.Sleep(SleepDuration)

				// проверяем, есть ли еще активные задачи
				if atomic.LoadInt32(&e.activeTasks) == 0 {
					log.Println("Активных задач нет")
					cancel()
					return
				}
			}
		}
	}
}
