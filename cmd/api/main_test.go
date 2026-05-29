package main

import "testing"

func TestMainEntrypointIsLinked(t *testing.T) {
	entrypoint := main
	if entrypoint == nil {
		t.Fatalf("expected main entrypoint to be linked")
	}
}
