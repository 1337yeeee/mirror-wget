package normalizer

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

type Normalizer interface {
	Normalize(baseURL, ref string) (string, error)
	SavePath() ([]string, error)
}

type DefaultNormalizer struct {
	URL *url.URL
}

func New() *DefaultNormalizer {
	return &DefaultNormalizer{}
}

func (n *DefaultNormalizer) Normalize(baseURL, ref string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u, err := base.Parse(ref)
	if err != nil {
		return "", err
	}
	u.Fragment = "" // отбрасываем #anchor
	// приводим схему и хост к нижнему регистру
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	// удаляем лишние слэши
	u.Path = path.Clean(u.Path)

	n.URL = u

	return u.String(), nil
}

func (n *DefaultNormalizer) SavePath() ([]string, error) {
	if n.URL == nil {
		return nil, errors.New("empty URL")
	}

	var savePath []string

	for _, part := range strings.Split(n.URL.Host+n.URL.Path, "/") {
		savePath = append(savePath, part)
	}

	return savePath, nil
}
