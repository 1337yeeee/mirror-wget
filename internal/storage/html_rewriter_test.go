package storage

import (
	"bytes"
	"context"
	"golang.org/x/net/html"
	"io"
	"mirror-wget/internal/normalizer"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// normalizeHTML используется для сравнения html документов
func normalizeHTML(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func _makeTestFile(fpath string, data []byte, t *testing.T) {
	t.Helper()
	dir := filepath.Dir(fpath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		t.Fatal(err)
	}
}

func _rmTestFile(fpath string, t *testing.T) {
	t.Helper()
	if err := os.Remove(fpath); err != nil {
		t.Fatal(err)
	}
}

// site map tree
// см. твой initDownLoadMap — оставляем как есть
func initDownLoadMap() *sync.Map {
	var m sync.Map
	m.Store("http://localhost:8080/", "localhost:8080/index.html")
	m.Store("http://localhost:8080/page2.html", "localhost:8080/page2.html")
	m.Store("http://localhost:8080/page3.html", "localhost:8080/page3.html")
	m.Store("http://localhost:8080/style.css", "localhost:8080/style.css")
	m.Store("http://localhost:8080/lvl1/", "localhost:8080/lvl1/index.html")
	m.Store("http://localhost:8080/lvl1/bar.html", "localhost:8080/lvl1/bar.html")
	m.Store("http://localhost:8080/lvl1/foo.html", "localhost:8080/lvl1/foo.html")
	m.Store("http://localhost:8080/lvl1/decor.css", "localhost:8080/lvl1/decor.css")
	m.Store("http://localhost:8080/lvl1/style2.css", "localhost:8080/lvl1/style2.css")
	m.Store("http://localhost:8080/lvl1/lvl2/", "localhost:8080/lvl1/lvl2/index.html")
	m.Store("http://localhost:8080/lvl1/lvl2/eggs.html", "localhost:8080/lvl1/lvl2/eggs.html")
	m.Store("http://localhost:8080/lvl1/lvl2/span.html", "localhost:8080/lvl1/lvl2/span.html")
	m.Store("http://localhost:8080/lvl1/lvl2/style3.css", "localhost:8080/lvl1/lvl2/style3.css")
	return &m
}

func runHTMLRewriteTest(t *testing.T, currentDocURL string, data, expect string) {
	docURL, err := normalizer.NewNormalizedUrl(currentDocURL)
	if err != nil {
		t.Fatal(err)
	}
	downloadMap := initDownLoadMap()
	pathResolver := NewPathResolver(docURL, downloadMap)
	rewriter := NewHTMLRewriter(pathResolver)

	filePath, err := docURL.SavePath()
	if err != nil {
		t.Fatal(err)
	}
	_makeTestFile(filePath, []byte(data), t)
	defer _rmTestFile(filePath, t)

	if err := rewriter.Rewrite(context.Background(), filePath); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}

	normGot, err := normalizeHTML(string(content))
	if err != nil {
		t.Fatal(err)
	}
	normExpected, err := normalizeHTML(expect)
	if err != nil {
		t.Fatal(err)
	}

	if normGot != normExpected {
		t.Errorf("HTML content mismatch\nGot:\n%s\n\nExpected:\n%s", normGot, normExpected)
	}
}

func TestHTMLRewrite(t *testing.T) {
	tests := []struct {
		name   string
		docURL string
		input  string
		expect string
	}{
		//            <a href="http://localhost:8080/index.html">Go home</a>
		//            <a href="../index.html">Go home</a>
		{
			name:   "basic links and css",
			docURL: "http://localhost:8080/lvl1/index.html",
			input: `
<html>
  <head>
    <link rel="stylesheet" href="http://localhost:8080/lvl1/lvl2/style3.css">
  </head>
  <body>
    <h1>Main page</h1>
    <a href="http://localhost:8080/index.html">Go home</a>
    <img src="foo.html">
    <a href="../page3.html">Go to page 3</a>
    <a href="../page100000000.html">Go to nowhere</a>
    <a href="lvl2/eggs.html">Go to eggs</a>
    <a href="http://localhost:8080/notexists/index.html">Broken</a>
  </body>
</html>`,
			expect: `
<html>
  <head>
    <link rel="stylesheet" href="lvl2/style3.css">
  </head>
  <body>
    <h1>Main page</h1>
    <a href="../index.html">Go home</a>
    <img src="foo.html">
    <a href="../page3.html">Go to page 3</a>
    <a href="http://localhost:8080/page100000000.html">Go to nowhere</a>
    <a href="lvl2/eggs.html">Go to eggs</a>
    <a href="http://localhost:8080/notexists/index.html">Broken</a>
  </body>
</html>`,
		},
		{
			name:   "script and anchor links",
			docURL: "http://localhost:8080/index.html",
			input: `
<html>
  <body>
    <script src="page2.html"></script>
    <a href="#section">Anchor</a>
	<a href="../#section2">Anchor Relative</a>
    <a href="mailto:test@example.com">Mail</a>
    <a href="/page3.html">Abs path</a>
  </body>
</html>`,
			expect: `
<html>
  <body>
    <script src="page2.html"></script>
    <a href="#section">Anchor</a>
	<a href="index.html#section2">Anchor Relative</a>
    <a href="mailto:test@example.com">Mail</a>
    <a href="page3.html">Abs path</a>
  </body>
</html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runHTMLRewriteTest(t, tt.docURL, tt.input, tt.expect)
		})
	}
}
