package engine

import (
	"context"
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

type Engine struct {
	baseURL     *normalizer.NormalizedUrl
	queue       queue.Queue
	visited     sync.Map
	numWorkers  int
	maxDepth    int
	wg          *sync.WaitGroup
	activeTasks int32
	workers     []*Worker
	robotsTxt   *downloader.Robots
}

func NewEngine(URL *normalizer.NormalizedUrl, robotsTxt *downloader.Robots, numWorkers, maxDepth int) *Engine {
	return &Engine{
		baseURL:    URL,
		queue:      queue.NewQueue(),
		visited:    sync.Map{},
		numWorkers: numWorkers,
		maxDepth:   maxDepth,
		wg:         &sync.WaitGroup{},
		robotsTxt:  robotsTxt,
	}
}

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

	engine := NewEngine(normURL, robotsTxt, runtime.GOMAXPROCS(0)-1, config.Level)
	return engine.Start()
}

func (e *Engine) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan queue.Item, 1024)
	item := queue.Item{
		URL:   e.baseURL,
		Depth: 0,
	}
	jobs <- item
	atomic.AddInt32(&e.activeTasks, 1)

	for n := 0; n < e.numWorkers; n++ {
		e.wg.Add(1)
		w := NewWorker(e.baseURL, e.wg, &e.activeTasks, e.queue)
		e.workers = append(e.workers, w)
		go w.worker(ctx, n, jobs)
	}

	e.wg.Add(1)
	go e.dispatcher(ctx, jobs, cancel)

	log.Println("Ждем завершения го рутин")
	e.wg.Wait()

	return nil
}

func (e *Engine) dispatcher(ctx context.Context, jobs chan<- queue.Item, cancel context.CancelFunc) {
	defer e.wg.Done()
	defer close(jobs)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if item, ok := e.queue.Pop(); ok {
				if e.maxDepth >= 0 && item.Depth > e.maxDepth {
					atomic.AddInt32(&e.activeTasks, -1)
					continue
				}
				if _, visited := e.visited.LoadOrStore(item.URL, true); visited {
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
