package parser

import "io"

type LinkParser interface {
	Parse(r io.Reader) error
	GetLinks() []string
}
