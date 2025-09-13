package parser

import (
	"strings"
	"testing"
)

func TestHTMLParser_Parse(t *testing.T) {
	tests := []struct {
		name   string
		html   string
		expect []string
	}{
		{
			name:   "empty html",
			html:   "",
			expect: []string{},
		},
		{
			name:   "no anchor tags",
			html:   "<p>Hello world</p><div>Nothing here</div>",
			expect: []string{},
		},
		{
			name:   "single valid link",
			html:   `<a href="https://example.com">Example</a>`,
			expect: []string{"https://example.com"},
		},
		{
			name:   "two valid links",
			html:   `<a href="foo">Foo</a><a href="/bar/baz">BarBaz</a>`,
			expect: []string{"foo", "/bar/baz"},
		},
		{
			name:   "links with quotes and spaces",
			html:   `<a href=" https://google.com ">Google</a> <a href='https://bing.com'>Bing</a>`,
			expect: []string{" https://google.com ", "https://bing.com"},
		},
		{
			name:   "mixed case attribute",
			html:   `<A HREF="test.html">Test</A>`,
			expect: []string{"test.html"},
		},
		{
			name: "multiple links with extra whitespace and newlines",
			html: `
				<ul>
					<li><a href="/home">Home</a></li>
					<li><a href="/about">About</a></li>
				</ul>
			`,
			expect: []string{"/home", "/about"},
		},
		{
			name:   "invalid HTML: unclosed tags",
			html:   `<a href="one"><a href="two">`,
			expect: []string{"one", "two"},
		},
		{
			name:   "anchor without href (should be ignored)",
			html:   `<a name="top">Top</a><a href="valid">Valid</a>`,
			expect: []string{"valid"},
		},
		{
			name:   "anchor with empty href",
			html:   `<a href="">Empty</a><a href="#">Hash</a>`,
			expect: []string{"", "#"},
		},
		{
			name:   "anchor with javascript: URL",
			html:   `<a href="javascript:alert('hi')">Click</a>`,
			expect: []string{"javascript:alert('hi')"},
		},
		{
			name:   "anchor with mailto: URL",
			html:   `<a href="mailto:test@example.com">Email</a>`,
			expect: []string{"mailto:test@example.com"},
		},
		{
			name:   "anchor inside other elements",
			html:   `<div><span><a href="deep">Deep Link</a></span></div>`,
			expect: []string{"deep"},
		},
		{
			name:   "multiple anchors, one malformed",
			html:   `<a href="good">Good</a><a >broken</a><a href="also-good">Also Good</a>`,
			expect: []string{"good", "also-good"},
		},
		{
			name: "complex nested structure",
			html: `
				<nav>
					<a href="/">Home</a>
					<a href="/products">Products</a>
					<div>
						<a href="/contact">Contact</a>
					</div>
				</nav>
				<footer>
					<a href="/privacy">Privacy</a>
				</footer>
			`,
			expect: []string{"", "/products", "/contact", "/privacy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewHTMLParser()
			parser.Parse(strings.NewReader(tt.html))
			links := parser.GetLinks()

			// Compare lengths first for clearer failure messages
			if len(links) != len(tt.expect) {
				t.Errorf("Expected %d links, got %d", len(tt.expect), len(links))
			}

			// Check if all expected links are present (order doesn't matter)
			gotMap := make(map[string]bool)
			for _, link := range links {
				gotMap[link] = true
			}
			for _, exp := range tt.expect {
				if !gotMap[exp] {
					t.Errorf("Expected link %q not found in result", exp)
				}
			}

			// Optional: also check no extra links
			if len(gotMap) != len(tt.expect) {
				t.Errorf("Got extra links: %+v", links)
			}
		})
	}
}
