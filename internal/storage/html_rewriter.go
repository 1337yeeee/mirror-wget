package storage

import (
	"context"
	"golang.org/x/net/html"
	"os"
)

// HTMLRewriter структура для перезаписывателя HTML
type HTMLRewriter struct {
	pathResolver *PathResolver
}

// NewHTMLRewriter инициализирует HTMLRewriter
func NewHTMLRewriter(pathResolver *PathResolver) Rewriter {
	return &HTMLRewriter{pathResolver: pathResolver}
}

// Rewrite переписывает документ, заменяя ссылки на внешние ресурсы локальными
func (r *HTMLRewriter) Rewrite(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return err
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					link, shouldReplace := r.pathResolver.Resolve(attr.Val)
					if shouldReplace {
						n.Attr[i].Val = link
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	return html.Render(out, doc)
}
