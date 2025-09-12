package normalizer

import (
	"reflect"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		ref      string
		expected string
		savePath []string
	}{
		{
			name:     "absolute url https",
			base:     "https://Example.COM",
			ref:      "https://Example.COM/foo/bar",
			expected: "https://example.com/foo/bar",
			savePath: []string{"example.com", "foo", "bar"},
		},
		{
			name:     "relative path",
			base:     "https://example.com/dir/page.html",
			ref:      "sub/page2.html",
			expected: "https://example.com/dir/sub/page2.html",
			savePath: []string{"example.com", "dir", "sub", "page2.html"},
		},
		{
			name:     "root relative",
			base:     "https://example.com/dir/page.html",
			ref:      "/img/logo.png",
			expected: "https://example.com/img/logo.png",
			savePath: []string{"example.com", "img", "logo.png"},
		},
		{
			name:     "with fragment",
			base:     "https://example.com",
			ref:      "/index.html#section",
			expected: "https://example.com/index.html",
			savePath: []string{"example.com", "index.html"},
		},
		{
			name:     "normalize dot segments",
			base:     "https://example.com/a/b/",
			ref:      "../c/./d.html",
			expected: "https://example.com/a/c/d.html",
			savePath: []string{"example.com", "a", "c", "d.html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := New()

			got, err := n.Normalize(tt.base, tt.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}

			savePathGot, err := n.SavePath()
			if err != nil {
				t.Fatalf("unexpected error on savePath: %v", err)
			}
			if !reflect.DeepEqual(savePathGot, tt.savePath) {
				t.Errorf("save path expected %q, got %q", tt.savePath, savePathGot)
			}
		})
	}
}

func TestNormalizeInvalidBase(t *testing.T) {
	n := New()
	_, err := n.Normalize(":://bad_url", "/path")
	if err == nil {
		t.Error("expected error for invalid base url, got nil")
	}
}

func TestNormalizeInvalidRef(t *testing.T) {
	n := New()
	_, err := n.Normalize("https://example.com", "::bad_ref")
	if err == nil {
		t.Error("expected error for invalid ref url, got nil")
	}
}
