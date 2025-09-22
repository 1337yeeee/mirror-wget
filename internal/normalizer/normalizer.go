package normalizer

import (
	"errors"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

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

	// Убираем логику автоматического добавления слеша - это должно определяться сервером
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

	// если путь есть, но не заканчивается на "/", и расширения нет — считаем директорией
	if u.Path == "" || u.Path != "" && !strings.HasSuffix(u.Path, "/") && filepath.Ext(u.Path) == "" {
		u.Path += "/"
	}

	return &NormalizedUrl{URL: u}, nil
}

// String преобразование структуры в строку - ВАЖНО: не меняем оригинальный URL!
func (n *NormalizedUrl) String() string {
	str := n.URL.String()
	if strings.HasSuffix(str, "index.html") {
		str = strings.TrimSuffix(str, "index.html")
	}
	return str
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

	// Если путь пустой или заканчивается на / - это директория
	if p == "" || strings.HasSuffix(p, "/") {
		return filepath.Join(host, p, "index.html"), nil
	}

	// Если есть расширение - это файл
	ext := filepath.Ext(p)
	if ext != "" {
		return filepath.Join(host, p), nil
	}

	// Нет расширения - считаем директорией
	return filepath.Join(host, p, "index.html"), nil
}
