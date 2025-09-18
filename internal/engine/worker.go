package engine

import (
	"context"
	"fmt"
	"io"
	"log"
	"mirror-wget/internal/downloader"
	"mirror-wget/internal/normalizer"
	"mirror-wget/internal/parser"
	"mirror-wget/internal/queue"
	"mirror-wget/internal/storage"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	baseURL     *normalizer.NormalizedUrl
	URL         *normalizer.NormalizedUrl
	wg          *sync.WaitGroup
	activeTasks *int32
	queue       queue.Queue
	stage       int
}

func NewWorker(
	baseURL *normalizer.NormalizedUrl,
	wg *sync.WaitGroup,
	activeTasks *int32,
	queue queue.Queue) *Worker {
	return &Worker{
		baseURL:     baseURL,
		URL:         baseURL,
		wg:          wg,
		activeTasks: activeTasks,
		queue:       queue,
	}
}

func (w *Worker) worker(ctx context.Context, id int, jobs <-chan queue.Item) {
	defer w.wg.Done()

	log.Printf("worker %d starting\n", id)

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

func (w *Worker) processItem(ctx context.Context, item queue.Item) {
	log.Printf("Processing: %s (depth: %d)\n", item.URL, item.Depth)
	w.URL = item.URL
	w.stage = 0

	var content []byte
	var contentType string
	var err error
	var links []string

	content, contentType, err = w.downloadFile(ctx, item)
	if err != nil {
		log.Printf("download file error: %s\n", err)
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
		err = w.saveFile(content, item)
		if err != nil {
			log.Printf("save file error: %s\n", err)
			return
		}
	}

	select {
	case <-ctx.Done():
		return
	default:
		links, err = w.parseFile(content, contentType, item)
		if err != nil {
			log.Printf("parse file error: %s\n", err)
			return
		}

	}

	select {
	case <-ctx.Done():
		return
	default:
		w.handleLinks(links, item.Depth)
	}
}

func (w *Worker) handleLinks(links []string, depth int) {
	for _, link := range links {
		newNorm, err := w.baseURL.Normalize(link)
		if err != nil {
			log.Printf("Normalize failed: %s - %v\n", link, err)
			continue
		}

		if newNorm.GetHost() == w.baseURL.GetHost() {
			queueItem := queue.Item{
				URL:   newNorm,
				Depth: depth + 1,
			}
			ok := w.queue.Push(queueItem)
			if ok {
				atomic.AddInt32(w.activeTasks, 1)
			}
		}
	}
}

func (w *Worker) downloadFile(ctx context.Context, item queue.Item) ([]byte, string, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Скачиваем контент
	log.Printf("Downloading %s (depth: %d)\n", item.URL, item.Depth)
	content, contentType, err := downloader.Get(ctxWithTimeout, item.URL.String())
	if err != nil {
		return nil, contentType, fmt.Errorf("download failed: %s - %v", item.URL.String(), err)
	}
	log.Printf("Downloaded %s (contentType: %s)\n", item.URL.String(), contentType)

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, contentType, fmt.Errorf("download failed: %s - %v", item.URL.String(), err)
	}

	return contentBytes, contentType, nil
}

func (w *Worker) saveFile(content []byte, item queue.Item) error {
	filePath, err := item.URL.SavePath()
	if err != nil {
		return fmt.Errorf("save path failed: %s - %v", item.URL.String(), err)
	}

	log.Printf("Saving %s (filepath: %s, len: %d)\n", item.URL, filePath, len(content))
	n, err := storage.Save(filePath, content)
	if err != nil {
		return fmt.Errorf("save failed: %s - %v", filePath, err)
	}
	log.Printf("Saved %s (%d bytes)\n", item.URL, n)

	return nil
}

func (w *Worker) parseFile(content []byte, contentType string, item queue.Item) ([]string, error) {
	var p parser.LinkParser
	if downloader.IsHTML(contentType) {
		p = parser.NewHTMLParser()
	} else if downloader.IsCSS(contentType) {
		p = parser.NewCSSParser()
	} else {
		p = parser.NewDefaultParser()
	}

	log.Printf("Parsing %s\n", item.URL)
	err := p.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("parse failed: %s - %v", item.URL.String(), err)
	}

	links := p.GetLinks()
	log.Printf("Parsed %s, got %d links\n", item.URL.String(), len(links))

	return links, nil
}
