package engine

import (
	"context"
	"log"
	"mirror-wget/internal/normalizer"
	"mirror-wget/internal/queue"
	"mirror-wget/internal/storage"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// StorageWorker структура для задач по изменению интернет-ссылок на локальные
type StorageWorker struct {
	baseURL     *normalizer.NormalizedUrl
	wg          *sync.WaitGroup
	queue       queue.Queue
	activeTasks *int32
	downloadMap *sync.Map
}

// NewStorageWorker инициализирует StorageWorker
func NewStorageWorker(
	baseURL *normalizer.NormalizedUrl,
	wg *sync.WaitGroup,
	queue queue.Queue,
	activeTasks *int32,
	downloadMap *sync.Map) *StorageWorker {
	return &StorageWorker{
		baseURL:     baseURL,
		wg:          wg,
		queue:       queue,
		activeTasks: activeTasks,
		downloadMap: downloadMap,
	}
}

// Storage основной внешний метод StorageWorker, следит за контекстом и выполняет задачи
func (w *StorageWorker) Storage(ctx context.Context, id int, jobs <-chan queue.Item) {
	defer w.wg.Done()

	log.Printf("storage worker %d starting\n", id)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			item, ok := <-jobs
			if !ok {
				return
			}

			w.processItem(ctx, item)
			atomic.AddInt32(w.activeTasks, -1)
		}
	}
}

// processItem обработка задачи
func (w *StorageWorker) processItem(ctx context.Context, item queue.Item) {
	log.Printf("Processing %s\n", item.URL.String())
	fp, ok := w.downloadMap.Load(item.URL.String())
	if !ok {
		return
	}
	log.Printf("Loaded: %s (filepath: %s)\n", item.URL.String(), fp)
	var st storage.Rewriter

	// which storage use
	pathResolver := storage.NewPathResolver(item.URL, w.downloadMap)
	extension := filepath.Ext(fp.(string))
	if extension == "html" {
		st = storage.NewHTMLRewriter(pathResolver)
	} else if extension == "css" {
		log.Printf("Loaded: %s (filepath: %s)\n", item.URL.String(), fp)
		st = storage.NewCSSRewriter(pathResolver)
	} else {
		return
	}

	err := st.Rewrite(ctx, fp.(string))
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Rewritten: %s\n", item.URL.String())
}
