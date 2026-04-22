package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"halalbrowse-ai/core/blocklist"
	"halalbrowse-ai/core/ml"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		configPath := serveCmd.String("config", defaultConfigPath(), "path to config.yaml")
		serveCmd.Parse(os.Args[2:])
		cfg, err := LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
		proxy, err := NewProxy(cfg)
		if err != nil {
			log.Fatalf("new proxy: %v", err)
		}
		if err := proxy.StartWithReload(*configPath); err != nil {
			log.Fatalf("proxy stopped: %v", err)
		}
	case "blocklist":
		blocklistCmd := flag.NewFlagSet("blocklist", flag.ExitOnError)
		configPath := blocklistCmd.String("config", defaultConfigPath(), "path to config.yaml")
		blocklistCmd.Parse(os.Args[2:])
		args := blocklistCmd.Args()
		if len(args) == 0 || args[0] != "pull" {
			log.Fatalf("usage: halalbrowse blocklist pull --config path")
		}
		cfg, err := LoadConfig(*configPath)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
		if err := runBlocklistPull(cfg); err != nil {
			log.Fatalf("blocklist pull failed: %v", err)
		}
		fmt.Println("blocklist pull complete")
	case "score-text":
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("read stdin: %v", err)
		}
		score := ml.NewTextScorer().Score(string(payload))
		encoded, _ := json.MarshalIndent(score, "", "  ")
		fmt.Println(string(encoded))
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("HalalBrowse AI CLI")
	fmt.Println("Usage:")
	fmt.Println("  halalbrowse serve --config ~/.halalbrowse/config.yaml")
	fmt.Println("  halalbrowse blocklist pull --config ~/.halalbrowse/config.yaml")
	fmt.Println("  echo 'text' | halalbrowse score-text")
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(home, ".halalbrowse", "config.yaml")
}

func runBlocklistPull(cfg Config) error {
	request, err := http.NewRequest(http.MethodGet, cfg.Blocklist.SourceURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	syncClient := blocklist.NewSyncClient(cfg.Blocklist.SourceURL, cfg.Blocklist.SigningKey, cfg.Blocklist.CachePath, cfg.Blocklist.MetadataPath, nil)
	meta, err := syncClient.ReadMetadata()
	if err == nil && meta.ETag != "" {
		request.Header.Set("If-None-Match", meta.ETag)
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return syncClient.Pull(blocklist.FetchResult{StatusCode: resp.StatusCode, ETag: resp.Header.Get("ETag"), Body: body, Signature: resp.Header.Get("X-Blocklist-Signature")})
}

func scoreBody(textScorer ml.TextScorer, body []byte) (ml.Score, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ml.Score{Confidence: 0.5, Label: "halal", Reasons: []string{"empty body"}}, nil
	}
	return textScorer.Score(string(body)), nil
}
