package storage

import (
	"context"
	"mirror-wget/internal/normalizer"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func _makeCSSTestFile(fpath string, data []byte, t *testing.T) {
	t.Helper()
	dir := filepath.Dir(fpath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fpath, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func _rmCSSTestFile(fpath string, t *testing.T) {
	t.Helper()
	if err := os.Remove(fpath); err != nil {
		t.Fatal(err)
	}
}

func initCSSDownloadMap() *sync.Map {
	var m sync.Map
	// Добавляем правильные URL которые действительно будут скачаны
	m.Store("http://localhost:8080/assets/css/img/bg.png", "localhost:8080/assets/css/img/bg.png")
	m.Store("http://localhost:8080/assets/css/fonts/font.woff2", "localhost:8080/assets/css/fonts/font.woff2")
	m.Store("http://localhost:8080/assets/css/theme.css", "localhost:8080/assets/css/theme.css")
	return &m
}

func runCSSRewriteTest(t *testing.T, input, expect string) {
	docURL, err := normalizer.NewNormalizedUrl("http://localhost:8080/assets/css/style.css")
	if err != nil {
		t.Fatal(err)
	}
	downloadMap := initCSSDownloadMap()
	pathResolver := NewPathResolver(docURL, downloadMap)
	rewriter := NewCSSRewriter(pathResolver)

	filePath := "testdata/style.css"
	_makeCSSTestFile(filePath, []byte(input), t)
	defer _rmCSSTestFile(filePath, t)

	if err := rewriter.Rewrite(context.Background(), filePath); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != expect {
		t.Errorf("CSS rewrite mismatch\nGot:\n%s\n\nExpected:\n%s", got, expect)
	}
}

func TestCSSRewrite(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "basic url()",
			input: `
body {
  background: url("img/bg.png");
}`,
			expect: `
body {
  background: url("img/bg.png");
}`,
		},
		{
			name: "font-face with url",
			input: `
@font-face {
  src: url('fonts/font.woff2');
}`,
			expect: `
@font-face {
  src: url('fonts/font.woff2');
}`,
		},
		{
			name: "import css",
			input: `
@import "theme.css";
`,
			expect: `
@import "theme.css";
`,
		},
		{
			name: "leave untouched external link",
			input: `
@import url("https://cdn.example.com/reset.css");
`,
			expect: `
@import url("https://cdn.example.com/reset.css");
`,
		},
		{
			name: "url without quotes",
			input: `
body {
  background: url(img/bg.png);
}`,
			expect: `
body {
  background: url(img/bg.png);
}`,
		},
		{
			name: "absolute url",
			input: `
body {
	background: url("http://localhost:8080/assets/css/img/bg.png");
}`,
			expect: `
body {
	background: url("img/bg.png");
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCSSRewriteTest(t, tt.input, tt.expect)
		})
	}
}
