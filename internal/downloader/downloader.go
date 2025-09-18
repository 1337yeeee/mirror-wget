package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func Get(ctx context.Context, url string) (io.ReadCloser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	return resp.Body, resp.Header.Get("Content-Type"), nil
}

func IsHTML(contentType string) bool {
	return strings.HasPrefix(contentType, "text/html")
}

func IsCSS(contentType string) bool {
	return strings.HasPrefix(contentType, "text/css")
}
