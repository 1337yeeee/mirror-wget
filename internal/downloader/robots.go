package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/temoto/robotstxt"
)

const UserAgent = "mirror-wget/0.1 (+https://github.com/yourname/mirror-wget)"

// Robots структура для работы с robots.txt
type Robots struct {
	data *robotstxt.RobotsData
}

// LoadRobots загружает robots.txt для данного базового URL.
func LoadRobots(base *url.URL) (*Robots, error) {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", base.Scheme, base.Host)
	resp, err := http.Get(robotsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		rdata, err := robotstxt.FromBytes(body)
		if err != nil {
			return nil, err
		}
		return &Robots{data: rdata}, nil
	}

	// Если файл не найден, или запрещён, или ошибка — политика по умолчанию
	// Например, при 404 считаем всё разрешённым
	return &Robots{data: nil}, nil
}

// Allowed проверяет путь URL — разрешён ли он.
func (r *Robots) Allowed(u *url.URL) bool {
	if r == nil || r.data == nil {
		return true
	}
	return r.data.TestAgent(u.Path, UserAgent)
}
