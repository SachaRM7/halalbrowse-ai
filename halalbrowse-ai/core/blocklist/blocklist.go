package blocklist

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Blocklist struct {
	Version       int      `json:"version"`
	GeneratedAt   string   `json:"generated_at,omitempty"`
	Domains       []string `json:"domains,omitempty"`
	URLPatterns   []string `json:"url_patterns,omitempty"`
	ContentHashes []string `json:"content_hashes,omitempty"`
}

func (b Blocklist) Matches(rawURL string) bool {
	for _, pattern := range b.Domains {
		if matchDomain(pattern, rawURL) {
			return true
		}
	}
	for _, pattern := range b.URLPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		if re.MatchString(rawURL) {
			return true
		}
	}
	return false
}

func (b Blocklist) HasContentHash(hash string) bool {
	hash = strings.ToLower(strings.TrimSpace(hash))
	for _, item := range b.ContentHashes {
		if strings.ToLower(item) == hash {
			return true
		}
	}
	return false
}

func LoadSigned(payload []byte, signatureHex string, secret string) (Blocklist, error) {
	if err := VerifySignature(payload, signatureHex, secret); err != nil {
		return Blocklist{}, err
	}
	var list Blocklist
	if err := json.Unmarshal(payload, &list); err != nil {
		return Blocklist{}, fmt.Errorf("decode blocklist: %w", err)
	}
	if list.Version <= 0 {
		return Blocklist{}, errors.New("blocklist version must be positive")
	}
	return list, nil
}

func VerifySignature(payload []byte, signatureHex string, secret string) error {
	decoded, err := hex.DecodeString(strings.TrimSpace(signatureHex))
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	if !hmac.Equal(decoded, mac.Sum(nil)) {
		return errors.New("invalid blocklist signature")
	}
	return nil
}

func matchDomain(pattern, rawURL string) bool {
	host := rawURL
	if strings.Contains(rawURL, "://") {
		parts := strings.SplitN(rawURL, "://", 2)
		host = parts[1]
	}
	host = strings.Split(host, "/")[0]
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	host = strings.ToLower(strings.TrimSpace(host))
	if pattern == "" || host == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*.")
		return host == suffix || strings.HasSuffix(host, "."+suffix)
	}
	return host == pattern
}
