package lzstring

import (
	"testing"
)

func TestLZStringDecompress(t *testing.T) {
	// 'BYUwNmD2AEDukCcwBMg=' is 'hello world'
	compressed := "BYUwNmD2AEDukCcwBMg="
	decompressed, err := DecompressFromBase64(compressed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decompressed != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", decompressed)
	}
}
