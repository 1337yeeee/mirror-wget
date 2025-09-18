package parser

import "io"

// DefaultParser реализует интерфейс LinkParser, возвращает пустой срез
type DefaultParser struct {
}

// NewDefaultParser инициализирует DefaultParser
func NewDefaultParser() LinkParser {
	return &DefaultParser{}
}

func (d DefaultParser) Parse(_ io.Reader) error {
	return nil
}

func (d DefaultParser) GetLinks() []string {
	return []string{}
}
