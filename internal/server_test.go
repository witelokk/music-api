package internal

import (
	"context"
	"testing"

	openapi "github.com/witelokk/music-api/internal/openapi"
)

func TestServer_GetHealth(t *testing.T) {
	server := &Server{}

	resp, err := server.GetHealth(context.Background(), openapi.GetHealthRequestObject{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	okResp, ok := resp.(openapi.GetHealth204Response)
	if !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}

	_ = okResp
}
