package blocklist

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type FetchResult struct {
	StatusCode int
	ETag       string
	Body       []byte
	Signature  string
}

type Metadata struct {
	ETag      string    `json:"etag"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	SourceURL string    `json:"source_url"`
}

type SyncClient struct {
	SourceURL string
	Secret    string
	CachePath string
	MetaPath  string
	Now       func() time.Time
}

func NewSyncClient(sourceURL, secret, cachePath, metaPath string, now func() time.Time) SyncClient {
	if now == nil {
		now = time.Now
	}
	return SyncClient{SourceURL: sourceURL, Secret: secret, CachePath: cachePath, MetaPath: metaPath, Now: now}
}

func (c SyncClient) Pull(result FetchResult) error {
	if result.StatusCode == 304 {
		return nil
	}
	if result.StatusCode != 200 {
		return fmt.Errorf("unexpected sync status: %d", result.StatusCode)
	}
	list, err := LoadSigned(result.Body, result.Signature, c.Secret)
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.CachePath, result.Body, 0o644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	meta := Metadata{ETag: result.ETag, Version: list.Version, UpdatedAt: c.Now().UTC(), SourceURL: c.SourceURL}
	encoded, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	if err := os.WriteFile(c.MetaPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}

func (c SyncClient) ReadMetadata() (Metadata, error) {
	content, err := os.ReadFile(c.MetaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Metadata{}, nil
		}
		return Metadata{}, err
	}
	var meta Metadata
	if err := json.Unmarshal(content, &meta); err != nil {
		return Metadata{}, err
	}
	return meta, nil
}

func (c SyncClient) LoadCached() (Blocklist, error) {
	content, err := os.ReadFile(c.CachePath)
	if err != nil {
		return Blocklist{}, err
	}
	var list Blocklist
	if err := json.Unmarshal(content, &list); err != nil {
		return Blocklist{}, err
	}
	return list, nil
}
