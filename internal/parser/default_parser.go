package parser

import "io"

// DefaultParser реализует интерфейс LinkParser, возвращает пустой срез
type DefaultParser struct {
}

// NewDefaultParser инициализирует DefaultParser
func NewDefaultParser() LinkParser {
	return &DefaultParser{}
}

// Parse ничего не делает
func (d DefaultParser) Parse(_ io.Reader) error {
	return nil
}

// GetLinks ничего не делает
func (d DefaultParser) GetLinks() []string {
	return []string{}
}
