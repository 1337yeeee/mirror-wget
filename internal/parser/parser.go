package parser

import "io"

// LinkParser интерфейс парсера ссылок
type LinkParser interface {
	Parse(r io.Reader) error
	GetLinks() []string
}
