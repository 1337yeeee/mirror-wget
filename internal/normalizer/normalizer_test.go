package normalizer

import (
	"reflect"
	"testing"
)

// TestNormalize тест нормализованного пути и пути сохранения
func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		ref      string
		expected string
		savePath string
	}{
		{
			name:     "absolute url https",
			base:     "https://Example.COM",
			ref:      "https://Example.COM/foo/bar",
			expected: "https://example.com/foo/bar/",
			savePath: "example.com/foo/bar/index.html",
		},
		{
			name:     "relative path",
			base:     "https://example.com/dir/page.html",
			ref:      "sub/page2.html",
			expected: "https://example.com/dir/sub/page2.html",
			savePath: "example.com/dir/sub/page2.html",
		},
		{
			name:     "root relative",
			base:     "https://example.com/dir/page.html",
			ref:      "/img/logo.png",
			expected: "https://example.com/img/logo.png",
			savePath: "example.com/img/logo.png",
		},
		{
			name:     "with fragment",
			base:     "https://example.com",
			ref:      "/index.html#section",
			expected: "https://example.com/",
			savePath: "example.com/index.html",
		},
		{
			name:     "normalize dot segments",
			base:     "https://example.com/a/b/",
			ref:      "../c/./d.html",
			expected: "https://example.com/a/c/d.html",
			savePath: "example.com/a/c/d.html",
		},
		{
			name:     "host",
			base:     "https://example.com",
			ref:      "",
			expected: "https://example.com/",
			savePath: "example.com/index.html",
		},
		{
			name:     "host with slash",
			base:     "https://example.com/",
			ref:      "",
			expected: "https://example.com/",
			savePath: "example.com/index.html",
		},
		{
			name:     "css url",
			base:     "https://example.com/style.css",
			ref:      "",
			expected: "https://example.com/style.css",
			savePath: "example.com/style.css",
		},
		{
			name:     "ref local ./",
			base:     "http://localhost:8080/lvl1/lvl2/",
			ref:      "./eggs.html",
			expected: "http://localhost:8080/lvl1/lvl2/eggs.html",
			savePath: "localhost:8080/lvl1/lvl2/eggs.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := NewNormalizedUrl(tt.base)
			if err != nil {
				t.Fatal(err)
			}

			got, err := n.Normalize(tt.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}

			savePathGot, err := got.SavePath()
			if err != nil {
				t.Fatalf("unexpected error on savePath: %v", err)
			}
			if !reflect.DeepEqual(savePathGot, tt.savePath) {
				t.Errorf("save path expected %q, got %q", tt.savePath, savePathGot)
			}
		})
	}
}

// TestNormalizeInvalidBase тест ошибка невалидного url
func TestNormalizeInvalidBase(t *testing.T) {
	_, err := NewNormalizedUrl(":://bad_url")
	if err == nil {
		t.Error("expected error for invalid base url, got nil")
	}
}

// TestNormalizeInvalidRef тест невалидного относительного пути
func TestNormalizeInvalidRef(t *testing.T) {
	n, err := NewNormalizedUrl("https://example.com")
	if err != nil {
		t.Fatal(err)
	}

	_, err = n.Normalize("::bad_ref")
	if err == nil {
		t.Error("expected error for invalid ref url, got nil")
	}
}
