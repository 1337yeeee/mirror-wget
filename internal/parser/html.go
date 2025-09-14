package parser

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"strings"
)

// HREF_ATOMS представляет множество atom'ов HTML тегов, которые могут содержать валидный href атрибут
var HREF_ATOMS = map[atom.Atom]bool{
	atom.A:    true, // <a>
	atom.Area: true, // <area>
	atom.Base: true, // <base>
	atom.Link: true, // <link>
}

// SRC_ATOMS представляет множество atom'ов HTML тегов, которые могут содержать валидный src атрибут
var SRC_ATOMS = map[atom.Atom]bool{
	atom.Audio:  true, // <audio>
	atom.Embed:  true, // <embed>
	atom.Iframe: true, // <iframe>
	atom.Img:    true, // <img>
	atom.Script: true, // <script>
	atom.Source: true, // <source>
	atom.Track:  true, // <track>
	atom.Video:  true, // <video>
}

// HTMLParser представляет структуру, которая хранит ссылки, извлеченные из HTML
type HTMLParser struct {
	Links map[string]bool
}

// NewHTMLParser инициализирует HTMLParser
func NewHTMLParser() LinkParser {
	return &HTMLParser{
		Links: make(map[string]bool),
	}
}

// Parse парсит HTML-документ
func (p *HTMLParser) Parse(r io.Reader) error {
	tokenizer := html.NewTokenizer(r)
	tType := tokenizer.Next()

	for ; tType != html.ErrorToken; tType = tokenizer.Next() {
		if tType == html.StartTagToken {
			token := tokenizer.Token()
			if _, ok := HREF_ATOMS[token.DataAtom]; ok {
				p.extractAndAddLink(token, "href")
			} else if _, ok := SRC_ATOMS[token.DataAtom]; ok {
				p.extractAndAddLink(token, "src")
			}
		}
	}

	return tokenizer.Err()
}

// GetLinks возвращает слайс строк - ссылок, которые были найдены в HTML-документе
func (p *HTMLParser) GetLinks() []string {
	if len(p.Links) == 0 {
		return nil
	}

	links := make([]string, 0, len(p.Links))
	for link, _ := range p.Links {
		links = append(links, link)
	}
	return links
}

// extractAndAddLink извлекает ссылку из атрибута (attrKey) токена (token) и добавляет ее к множеству ссылок
func (p *HTMLParser) extractAndAddLink(token html.Token, attrKey string) {
	for _, attr := range token.Attr {
		if attr.Key == attrKey {
			link := strings.TrimRight(attr.Val, "/")
			link = strings.TrimSpace(link)
			p.Links[link] = true
			return
		}
	}
}
