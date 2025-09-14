package engine

import (
	"errors"
	"fmt"
	"mirror-wget/internal/cli"
	"mirror-wget/internal/downloader"
	"mirror-wget/internal/normalizer"
	"mirror-wget/internal/parser"
	"mirror-wget/internal/queue"
)

func Handle() error {
	q := queue.NewQueue()
	config, err := cli.NewConfig()
	if err != nil {
		return err
	}

	norm := normalizer.New()
	url, err := norm.Normalize(config.URL, "")
	if err != nil {
		return err
	}

	d := downloader.NewDownloader()
	html, err := d.Get(url)
	if err != nil {
		return err
	}

	var p parser.LinkParser
	if d.IsHTML() {
		p = parser.NewHTMLParser()
	} else if d.IsCSS() {
		p = parser.NewCSSParser()
	} else {
		return errors.New("unknown parser")
	}

	_ = p.Parse(html)
	links := p.GetLinks()
	for _, l := range links {
		err = q.Push(l)
	}

	fmt.Println(links)
	return nil
}

// алгоритм:
// скачиваем
// сохраняем
// парсим
// добавляем в очередь
// повторяем по очереди

// когда очередь пустая - постобработка файлов
