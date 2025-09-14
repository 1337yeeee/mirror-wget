package downloader

import (
	"io"
	"net/http"
)

type Downloader interface {
	Get(url string) (io.ReadCloser, error)
	IsHTML() bool
	IsCSS() bool
}

type Impl struct{}

func NewDownloader() Downloader {
	return &Impl{}
}

func (d *Impl) Get(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (d *Impl) IsHTML() bool {
	return true
}

func (d *Impl) IsCSS() bool {
	return true
}
