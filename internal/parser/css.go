package parser

import (
	"errors"
	"github.com/riking/cssparse/tokenizer"
	"io"
	"strings"
)

// CSSParser представляет структуру, которая хранит ссылки, извлеченные из CSS
type CSSParser struct {
	Links map[string]bool
}

// NewCSSParser инициализирует CSSParser
func NewCSSParser() LinkParser {
	return &CSSParser{
		Links: make(map[string]bool),
	}
}

// Parse парсит HTML-документ
func (p *CSSParser) Parse(r io.Reader) error {
	t := tokenizer.NewTokenizer(r)
	for {
		token := t.Next()
		if token.Type.StopToken() {
			break
		}

		if token.Type == tokenizer.TokenAtKeyword && strings.ToLower(token.Value) == "import" {
			link, err := p.scanImport(t)
			if err != nil {
				return err
			}
			p.postProcessAndAddLink(link)
		} else if token.Type == tokenizer.TokenURI {
			p.postProcessAndAddLink(token.Value)
		}
	}
	return nil
}

// GetLinks возвращает слайс строк - ссылок, которые были найдены в HTML-документе
func (p *CSSParser) GetLinks() []string {
	if len(p.Links) == 0 {
		return nil
	}

	links := make([]string, 0, len(p.Links))
	for link := range p.Links {
		links = append(links, link)
	}
	return links
}

// postProcessAndAddLink обрабатывает ссылку и добавляет ее к множеству ссылок
func (p *CSSParser) postProcessAndAddLink(link string) {
	link = strings.TrimRight(link, "/")
	link = strings.TrimSpace(link)
	p.Links[link] = true
}

// scanImport обрабатывает ключевое слово `import`
// Ниже представлен Formal syntax
// @import =
//
//	@import [ <url> | <string> ] [ layer | layer( <layer-name> ) ]? <import-conditions> ;
//
// <url> =
//
//	<url()>  |
//	<src()>  # включает в себя вызов url()
func (p *CSSParser) scanImport(t *tokenizer.Tokenizer) (string, error) {
	for token := t.Next(); token.Type != tokenizer.TokenSemicolon || token.Type.StopToken(); token = t.Next() {
		if token.Type == tokenizer.TokenString || token.Type == tokenizer.TokenURI {
			return token.Value, nil
		}
	}

	return "", errors.New("unexpected token")
}
