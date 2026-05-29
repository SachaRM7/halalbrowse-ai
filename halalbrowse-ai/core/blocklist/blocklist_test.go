package blocklist

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func signedPayload(t *testing.T, secret string, list Blocklist) ([]byte, string) {
	t.Helper()
	payload, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return payload, hex.EncodeToString(mac.Sum(nil))
}

func TestLoadSignedBlocklistVerifiesSignatureAndMatches(t *testing.T) {
	expected := Blocklist{
		Version:       3,
		GeneratedAt:   "2026-04-22T22:00:00Z",
		Domains:       []string{"*.bad.example", "casino.example"},
		URLPatterns:   []string{"https://bad.example/.*"},
		ContentHashes: []string{"abc123"},
	}
	payload, signature := signedPayload(t, "secret", expected)

	loaded, err := LoadSigned(payload, signature, "secret")
	if err != nil {
		t.Fatalf("LoadSigned returned error: %v", err)
	}

	if !loaded.Matches("https://shop.bad.example/product") {
		t.Fatalf("expected wildcard domain match")
	}
	if !loaded.Matches("https://bad.example/gallery/123") {
		t.Fatalf("expected regex URL match")
	}
	if loaded.Matches("https://halal.example") {
		t.Fatalf("did not expect allow URL to match")
	}
}

func TestSyncPullWritesConditionalCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "blocklist.json")
	metaPath := filepath.Join(dir, "blocklist.meta.json")
	expected := Blocklist{Version: 7, Domains: []string{"example.com"}}
	payload, signature := signedPayload(t, "secret", expected)

	client := NewSyncClient("https://example.com/blocklist.json", "secret", cachePath, metaPath, nil)
	err := client.Pull(FetchResult{StatusCode: 200, ETag: "etag-1", Body: payload, Signature: signature})
	if err != nil {
		t.Fatalf("Pull returned error: %v", err)
	}

	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to exist: %v", err)
	}
	meta, err := client.ReadMetadata()
	if err != nil {
		t.Fatalf("ReadMetadata returned error: %v", err)
	}
	if meta.ETag != "etag-1" {
		t.Fatalf("expected ETag to be persisted, got %q", meta.ETag)
	}
}
