package models_test

import (
	"testing"

	"github.com/2Geigh/Ohana/db-init/internal/models"
)

func TestModels_LocalQueueDequeue_EmptyQueue(t *testing.T) {
	q := models.LocalQueue{}

	got := q.Dequeue()

	if got != models.Url("") {
		t.Fatalf("Dequeue() = %q, want empty models.Url", got)
	}

	if len(q.Links) != 0 {
		t.Fatalf("queue length = %d, want 0", len(q.Links))
	}
}

func TestModels_LocalQueueDequeue_SingleItem(t *testing.T) {
	q := models.LocalQueue{
		Links: []models.Url{"https://example.com"},
	}

	got := q.Dequeue()

	if got != models.Url("https://example.com") {
		t.Fatalf("Dequeue() = %q, want %q", got, models.Url("https://example.com"))
	}

	if len(q.Links) != 0 {
		t.Fatalf("queue length = %d, want 0", len(q.Links))
	}
}

func TestModels_LocalQueueDequeue_RemovesItemsInFIFOOrder(t *testing.T) {
	q := models.LocalQueue{
		Links: []models.Url{
			"https://example.com/1",
			"https://example.com/2",
			"https://example.com/3",
		},
	}

	tests := []struct {
		name string
		want models.Url
	}{
		{
			name: "first item",
			want: "https://example.com/1",
		},
		{
			name: "second item",
			want: "https://example.com/2",
		},
		{
			name: "third item",
			want: "https://example.com/3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := q.Dequeue()

			if got != tt.want {
				t.Fatalf("Dequeue() = %q, want %q", got, tt.want)
			}
		})
	}

	if len(q.Links) != 0 {
		t.Fatalf("queue length = %d, want 0", len(q.Links))
	}
}

func TestModels_LocalQueueDequeue_AfterQueueIsEmpty(t *testing.T) {
	q := models.LocalQueue{
		Links: []models.Url{"https://example.com"},
	}

	_ = q.Dequeue()

	got := q.Dequeue()

	if got != models.Url("") {
		t.Fatalf("Dequeue() after emptying queue = %q, want empty models.Url", got)
	}
}
