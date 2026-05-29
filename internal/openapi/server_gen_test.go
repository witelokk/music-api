package openapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratedEnumValidation(t *testing.T) {
	if !Album.Valid() {
		t.Fatalf("expected album release type to be valid")
	}
	if ReleaseType("mixtape").Valid() {
		t.Fatalf("expected unknown release type to be invalid")
	}
	if !SongPlay.Valid() {
		t.Fatalf("expected song_play event type to be valid")
	}
	if SearchResultItemType("podcast").Valid() {
		t.Fatalf("expected unknown search result type to be invalid")
	}
}

func TestGeneratedMediaResponseWritesHeadersAndBody(t *testing.T) {
	w := httptest.NewRecorder()
	resp := GetMedia200ImagejpegResponse{
		Body:          strings.NewReader("image"),
		ContentLength: 5,
	}

	if err := resp.VisitGetMediaResponse(w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("expected content type %q, got %q", "image/jpeg", got)
	}
	if got := w.Header().Get("Content-Length"); got != "5" {
		t.Fatalf("expected content length %q, got %q", "5", got)
	}
	if got := w.Body.String(); got != "image" {
		t.Fatalf("expected body %q, got %q", "image", got)
	}
}
