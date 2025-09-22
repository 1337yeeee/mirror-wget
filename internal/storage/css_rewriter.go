package storage

import (
	"context"
	"os"
	"regexp"
	"strings"
)

// CSSRewriter структура для перезаписывателя CSS
type CSSRewriter struct {
	pathResolver *PathResolver
}

// NewCSSRewriter инициализирует CSSRewriter
func NewCSSRewriter(pathResolver *PathResolver) Rewriter {
	return &CSSRewriter{pathResolver: pathResolver}
}

var (
	reURL    = regexp.MustCompile(`url\(([^)]+)\)`)
	reImport = regexp.MustCompile(`@import\s+(?:url\()?['"]?([^'")]+)['"]?\)?`)
)

// Rewrite переписывает документ, заменяя ссылки на внешние ресурсы локальными
func (r *CSSRewriter) Rewrite(ctx context.Context, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	css := string(data)

	// Функция для удаления обрамляющих кавычек
	trimQuotes := func(s string) string {
		s = strings.TrimSpace(s)
		if len(s) >= 2 {
			if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
				return s[1 : len(s)-1]
			}
		}
		return s
	}

	// Обработчик замены
	replacer := func(match, sub string) string {
		cleanSub := trimQuotes(sub)
		resolved, ok := r.pathResolver.Resolve(cleanSub)
		if !ok {
			return match
		}

		// Определяем, какие кавычки использовать (если были)
		var quote string
		trimmed := strings.TrimSpace(sub)
		if len(trimmed) > 0 && (trimmed[0] == '"' || trimmed[0] == '\'') {
			quote = string(trimmed[0])
		}

		// Оборачиваем resolved в те же кавычки
		replacement := quote + resolved + quote

		// Заменяем только подстроку sub → replacement, сохраняя всю остальную структуру match
		return strings.Replace(match, sub, replacement, 1)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	css = reURL.ReplaceAllStringFunc(css, func(m string) string {
		sub := reURL.FindStringSubmatch(m)
		if len(sub) > 1 {
			return replacer(m, sub[1]) // sub[1] — содержимое внутри url(...)
		}
		return m
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	css = reImport.ReplaceAllStringFunc(css, func(m string) string {
		sub := reImport.FindStringSubmatch(m)
		if len(sub) > 1 {
			return replacer(m, sub[1]) // sub[1] — путь внутри @import
		}
		return m
	})

	return os.WriteFile(filePath, []byte(css), 0644)
}
