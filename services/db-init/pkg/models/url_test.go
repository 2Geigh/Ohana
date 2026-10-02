package models_test

import (
	"testing"

	"github.com/2Geigh/Ohana/db-init/pkg/models"
)

func TestUrl_TrimWWW(t *testing.T) {
	tests := []struct {
		name string
		url  models.Url
		want models.Url
	}{
		{
			name: "URL with protocol without www",
			url:  models.Url("gemini://api.example.com?foo=bar#section"),
			want: models.Url("gemini://api.example.com?foo=bar#section"),
		},
		{
			name: "URL without protocol without www",
			url:  models.Url("api.example.com?foo=bar#section"),
			want: models.Url("api.example.com?foo=bar#section"),
		},
		{
			name: "URL with protocol with www",
			url:  models.Url("gemini://www.api.example.com?foo=bar#section"),
			want: models.Url("gemini://api.example.com?foo=bar#section"),
		},
		{
			name: "URL without protocol with www",
			url:  models.Url("www.api.example.com?foo=bar#section"),
			want: models.Url("api.example.com?foo=bar#section"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.url.TrimWWW()

			if got != tt.want {
				t.Errorf(
					"GetSecondAndTopLevelDomain() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestUrl_GetDomain(t *testing.T) {
	tests := []struct {
		name string
		url  models.Url
		want models.Domain
	}{
		{
			name: "empty URL",
			url:  models.Url(""),
			want: models.Domain(""),
		},
		{
			name: "HTTPS YouTube URL",
			url:  models.Url("https://www.youtube.com/watch?v=dQw4w9WgXcQ"),
			want: models.Domain("youtube.com"),
		},
		{
			name: "HTTP example URL",
			url:  models.Url("http://www.example.com"),
			want: models.Domain("example.com"),
		},
		{
			name: "URL with a path",
			url:  models.Url("https://example.org/products/item"),
			want: models.Domain("example.org"),
		},
		{
			name: "URL with a query string",
			url:  models.Url("https://google.com/search?q=golang"),
			want: models.Domain("google.com"),
		},
		{
			name: "URL with a fragment",
			url:  models.Url("https://github.com/golang/go#readme"),
			want: models.Domain("github.com"),
		},
		{
			name: "URL with multiple subdomains",
			url:  models.Url("https://api.v1.example.com/users"),
			want: models.Domain("api.v1.example.com"),
		},
		{
			name: "WWW subdomain",
			url:  models.Url("https://www.wikipedia.org/wiki/Go"),
			want: models.Domain("wikipedia.org"),
		},
		{
			name: "Net top-level domain",
			url:  models.Url("https://www.example.net"),
			want: models.Domain("example.net"),
		},
		{
			name: "Educational domain",
			url:  models.Url("https://www.example.edu/courses"),
			want: models.Domain("example.edu"),
		},
		{
			name: "Government domain",
			url:  models.Url("https://www.example.gov/services"),
			want: models.Domain("example.gov"),
		},
		{
			name: "Country-code domain",
			url:  models.Url("https://www.example.co.uk"),
			want: models.Domain("example.co.uk"),
		},
		{
			name: "Country-code domain with subdomain",
			url:  models.Url("https://shop.example.co.uk/products"),
			want: models.Domain("shop.example.co.uk"),
		},
		{
			name: "Routeless url with query",
			url:  models.Url("https://support.github.com?tags=dotcom-footer"),
			want: models.Domain("support.github.com"),
		},
		{
			name: "URL without protocol",
			url:  models.Url("www.example.com/path"),
			want: models.Domain("example.com"),
		},
		{
			name: "Domain without protocol or www",
			url:  models.Url("example.com"),
			want: models.Domain("example.com"),
		},
		{
			name: "Routeless URL with a fragment",
			url:  models.Url("https://github.com#readme"),
			want: models.Domain("github.com"),
		},
		{
			name: "Routeless URL with query before fragment",
			url:  models.Url("https://example.com?foo=bar#section"),
			want: models.Domain("example.com"),
		},
		{
			name: "Routeless URL with fragment before query",
			url:  models.Url("https://example.com#section?foo=bar"),
			want: models.Domain("example.com"),
		},
		{
			name: "WWW URL with a query and no path",
			url:  models.Url("https://www.example.com?foo=bar"),
			want: models.Domain("example.com"),
		},
		{
			name: "WWW URL with a fragment and no path",
			url:  models.Url("https://www.example.com#section"),
			want: models.Domain("example.com"),
		},
		{
			name: "WWW URL with query before fragment and no path",
			url:  models.Url("https://www.example.com?foo=bar#section"),
			want: models.Domain("example.com"),
		},
		{
			name: "WWW URL with fragment before query and no path",
			url:  models.Url("https://www.example.com#section?foo=bar"),
			want: models.Domain("example.com"),
		},
		{
			name: "HTTP URL with query and no path",
			url:  models.Url("http://example.com?foo=bar"),
			want: models.Domain("example.com"),
		},
		{
			name: "HTTP URL with fragment and no path",
			url:  models.Url("http://example.com#section"),
			want: models.Domain("example.com"),
		},
		{
			name: "URL with a port",
			url:  models.Url("https://example.com:8080/path"),
			want: models.Domain("example.com:8080"),
		},
		{
			name: "WWW URL with a port",
			url:  models.Url("https://www.example.com:8080/path"),
			want: models.Domain("example.com:8080"),
		},
		{
			name: "URL with a subdomain, query, and fragment",
			url:  models.Url("https://api.example.com?foo=bar#section"),
			want: models.Domain("api.example.com"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.url.GetDomain()

			if got != tt.want {
				t.Errorf(
					"GetSecondAndTopLevelDomain() = %q, want %q",
					got,
					tt.want,
				)
			}
		})
	}
}
