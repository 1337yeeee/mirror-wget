package normalizer

import (
	"errors"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// Normalizer интерфейс нормализации URL
type Normalizer interface {
	Normalize(baseURL, ref string) (Normalizer, error)
	SavePath() (string, error)
	GetHost() string
	String() string
}

// NormalizedUrl структура для нормализации URL
type NormalizedUrl struct {
	URL *url.URL
}

// NewNormalizedUrl инициализация NormalizedUrl
func NewNormalizedUrl(baseURL string) (*NormalizedUrl, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	base.Fragment = ""
	base.Scheme = strings.ToLower(base.Scheme)
	base.Host = strings.ToLower(base.Host)

	return &NormalizedUrl{URL: base}, nil
}

// Normalize нормализация URL
func (n *NormalizedUrl) Normalize(ref string) (*NormalizedUrl, error) {
	u, err := n.URL.Parse(ref)
	if err != nil {
		return nil, err
	}

	u.Fragment = ""
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	if u.Host != n.URL.Host {
		return nil, errors.New("host does not match")
	}

	if u.Path != "" {
		u.Path = path.Clean(u.Path)
	}

	return &NormalizedUrl{URL: u}, nil
}

// String преобразование структуры в строку
func (n *NormalizedUrl) String() string {
	return n.URL.String()
}

// SavePath возвращает путь по которому нужно сохранить документ
func (n *NormalizedUrl) SavePath() (string, error) {
	return buildSavePath(n.URL)
}

// GetHost возвращает хост адреса
func (n *NormalizedUrl) GetHost() string {
	return n.URL.Host
}

// buildSavePath делает путь для сохранения
func buildSavePath(u *url.URL) (string, error) {
	host := u.Host
	p := u.Path

	// Если пусто или "/" -> index.html
	if p == "" || strings.HasSuffix(p, "/") {
		return filepath.Join(host, p, "index.html"), nil
	}

	ext := filepath.Ext(p)
	if ext == "" {
		// нет расширения, считаем что это html -> about/index.html
		return filepath.Join(host, p, "index.html"), nil
	}

	// есть расширение (css, js, png, html и т.д.)
	return filepath.Join(host, p), nil
}
