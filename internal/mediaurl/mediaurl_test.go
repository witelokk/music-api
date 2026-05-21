package mediaurl

import "testing"

func TestBuild_DefaultBasePath(t *testing.T) {
	SetBasePath("")

	if got := Build("stream-id"); got != "/api/v1/media/stream-id" {
		t.Fatalf("expected default media URL %q, got %q", "/api/v1/media/stream-id", got)
	}
}

func TestSetBasePath_PreservesLeadingSlash(t *testing.T) {
	SetBasePath("/media")

	if got := Build("stream-id"); got != "/media/stream-id" {
		t.Fatalf("expected media URL %q, got %q", "/media/stream-id", got)
	}
}

func TestSetBasePath_AbsoluteURL(t *testing.T) {
	SetBasePath("http://192.168.1.133/api/v1/media")

	if got := Build("stream-id"); got != "http://192.168.1.133/api/v1/media/stream-id" {
		t.Fatalf("expected media URL %q, got %q", "http://192.168.1.133/api/v1/media/stream-id", got)
	}
}

func TestSetBasePath_UpgradesLegacyAPIMediaPath(t *testing.T) {
	SetBasePath("http://192.168.1.133/api/media")

	if got := Build("stream-id"); got != "http://192.168.1.133/api/v1/media/stream-id" {
		t.Fatalf("expected media URL %q, got %q", "http://192.168.1.133/api/v1/media/stream-id", got)
	}
}
