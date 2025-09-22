package storage

import (
	"mirror-wget/internal/normalizer"
	"path/filepath"
	"strings"
	"sync"
)

// PathResolver отвечает за преобразование путей
type PathResolver struct {
	currentDocURL *normalizer.NormalizedURL
	downloadedMap *sync.Map
	baseHost      string
}

// NewPathResolver инициализирует PathResolver
func NewPathResolver(currentDocURL *normalizer.NormalizedURL, downloadedMap *sync.Map) *PathResolver {
	return &PathResolver{
		currentDocURL: currentDocURL,
		downloadedMap: downloadedMap,
		baseHost:      currentDocURL.GetHost(),
	}
}

// Resolve преобразует ссылку: возвращает новую ссылку и флаг нужно ли заменять
func (pr *PathResolver) Resolve(link string) (string, bool) {
	// запоминаем якорь
	var fragment string
	if idx := strings.Index(link, "#"); idx != -1 {
		if idx == 0 {
			return "", false
		}
		fragment = link[idx:]
		link = link[:idx]
	}

	// нормализуем ссылку относительно текущего документа
	normLink, err := pr.currentDocURL.Normalize(link)
	if err != nil {
		return link + fragment, false // не смогли нормализовать - оставляем как есть
	}

	// проверяем, скачан ли этот ресурс
	normLinkStr := normLink.String()
	if _, loaded := pr.downloadedMap.Load(normLinkStr); loaded {
		// ресурс скачан - делаем относительный путь

		relativePath, err := pr.makeRelativePath(normLink)
		if err != nil {
			return normLinkStr + fragment, true // если ошибка - возвращаем абсолютный путь
		}
		return relativePath + fragment, true
	}

	// ресурс НЕ скачан
	if pr.isAbsoluteURL(link) {
		// абсолютная ссылка на внешний ресурс - оставляем как есть
		return link + fragment, false
	}
	// относительная ссылка на нескачанный ресурс - делаем абсолютной
	return normLinkStr + fragment, true
}

// makeRelativePath создает относительный путь от текущего документа к целевому
func (pr *PathResolver) makeRelativePath(target *normalizer.NormalizedURL) (string, error) {
	currentPath, err := pr.currentDocURL.SavePath()
	if err != nil {
		return "", err
	}
	targetPath, err := target.SavePath()
	if err != nil {
		return "", err
	}

	// если пути одинаковые (ссылка на тот же файл) - возвращаем только имя файла
	if currentPath == targetPath {
		return filepath.Base(targetPath), nil
	}

	// получаем директории
	currentDir := filepath.Dir(currentPath)
	targetDir := filepath.Dir(targetPath)

	fileName := filepath.Base(targetPath)

	if currentDir == targetDir {
		return fileName, nil
	}

	// если в разных директориях - вычисляем относительный путь
	relPath, err := filepath.Rel(currentDir, targetPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

func (pr *PathResolver) isAbsoluteURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
