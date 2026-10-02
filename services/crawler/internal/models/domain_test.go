package models_test

import (
	"testing"

	"github.com/2Geigh/Ohana/crawler/internal/models"
)

func TestDomain_GetFQDN(t *testing.T) {
	tests := []struct {
		name  string
		input models.Domain
		want  models.Domain
	}{
		{
			name:  "HTTPS YouTube URL",
			input: models.Domain("https://www.youtube.com/watch?v=dQw4w9WgXcQ"),
			want:  models.Domain("youtube.com"),
		},
		{
			name:  "HTTP example URL",
			input: models.Domain("http://www.example.com"),
			want:  models.Domain("example.com"),
		},
		{
			name:  "URL with a path",
			input: models.Domain("https://example.org/products/item"),
			want:  models.Domain("example.org"),
		},
		{
			name:  "URL with a query string",
			input: models.Domain("https://google.com/search?q=golang"),
			want:  models.Domain("google.com"),
		},
		{
			name:  "URL with a fragment",
			input: models.Domain("https://github.com/golang/go#readme"),
			want:  models.Domain("github.com"),
		},
		{
			name:  "URL with multiple subdomains",
			input: models.Domain("https://api.v1.example.com/users"),
			want:  models.Domain("example.com"),
		},
		{
			name:  "WWW subdomain",
			input: models.Domain("https://www.wikipedia.org/wiki/Go"),
			want:  models.Domain("wikipedia.org"),
		},
		{
			name:  "Net top-level domain",
			input: models.Domain("https://www.example.net"),
			want:  models.Domain("example.net"),
		},
		{
			name:  "Educational domain",
			input: models.Domain("https://www.example.edu/courses"),
			want:  models.Domain("example.edu"),
		},
		{
			name:  "Government domain",
			input: models.Domain("https://www.example.gov/services"),
			want:  models.Domain("example.gov"),
		},
		{
			name:  "Country-code domain",
			input: models.Domain("https://www.example.co.uk"),
			want:  models.Domain("example.co.uk"),
		},
		{
			name:  "Country-code domain 1",
			input: models.Domain("https://www.tatoeba.co.jp"),
			want:  models.Domain("tatoeba.co.jp"),
		},
		{
			name:  "Country-code domain 2",
			input: models.Domain("https://www.amazon.com.cn"),
			want:  models.Domain("amazon.com.cn"),
		},
		{
			name:  "Country-code domain with subdomain",
			input: models.Domain("https://shop.example.co.uk/products"),
			want:  models.Domain("example.co.uk"),
		},
		{
			name:  "Empty URL",
			input: models.Domain(""),
			want:  models.Domain(""),
		},
		{
			name:  "Single-level domain",
			input: models.Domain("localhost"),
			want:  models.Domain("localhost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.GetFQDN()

			if got != tt.want {
				t.Errorf(
					"GetFQDN() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestDomain_StripProtocol(t *testing.T) {
	tests := []struct {
		name  string
		input models.Domain
		want  models.Domain
	}{
		{
			name:  "HTTPS URL",
			input: models.Domain("https://example.com"),
			want:  models.Domain("example.com"),
		},
		{
			name:  "HTTP URL",
			input: models.Domain("http://example.com"),
			want:  models.Domain("example.com"),
		},
		{
			name:  "HTTPS URL with path",
			input: models.Domain("https://example.com/products/item"),
			want:  models.Domain("example.com/products/item"),
		},
		{
			name:  "HTTP URL with query string",
			input: models.Domain("http://example.com/search?q=golang"),
			want:  models.Domain("example.com/search?q=golang"),
		},
		{
			name:  "URL with fragment",
			input: models.Domain("https://example.com#readme"),
			want:  models.Domain("example.com#readme"),
		},
		{
			name:  "URL with multiple subdomains",
			input: models.Domain("https://api.v1.example.com"),
			want:  models.Domain("api.v1.example.com"),
		},
		{
			name:  "Domain without protocol",
			input: models.Domain("example.com"),
			want:  models.Domain("example.com"),
		},
		{
			name:  "Domain without protocol and with path",
			input: models.Domain("example.com/products"),
			want:  models.Domain("example.com/products"),
		},
		{
			name:  "Empty domain",
			input: models.Domain(""),
			want:  models.Domain(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.StripProtocol()

			if got != tt.want {
				t.Errorf(
					"StripProtocol() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}
