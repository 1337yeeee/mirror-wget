package parser

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"strings"
)

// HrefAtoms представляет множество atom'ов HTML тегов, которые могут содержать валидный href атрибут
var HrefAtoms = map[atom.Atom]bool{
	atom.A:    true, // <a>
	atom.Area: true, // <area>
	atom.Base: true, // <base>
	atom.Link: true, // <link>
}

// SrcAtoms представляет множество atom'ов HTML тегов, которые могут содержать валидный src атрибут
var SrcAtoms = map[atom.Atom]bool{
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
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					p.extractAndAddLink(attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return nil
}

// GetLinks возвращает слайс строк - ссылок, которые были найдены в HTML-документе
func (p *HTMLParser) GetLinks() []string {
	if len(p.Links) == 0 {
		return nil
	}

	links := make([]string, 0, len(p.Links))
	for link := range p.Links {
		links = append(links, link)
	}
	return links
}

// extractAndAddLink добавляет ссылку к множеству ссылок
func (p *HTMLParser) extractAndAddLink(link string) {
	link = strings.TrimRight(link, "/")
	link = strings.TrimSpace(link)
	p.Links[link] = true
}
