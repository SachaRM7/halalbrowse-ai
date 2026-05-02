package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
	"halalbrowse-ai/core/blocklist"
	"halalbrowse-ai/core/ml"
	"halalbrowse-ai/core/prayertimes"
)

type Config struct {
	ListenAddr        string  `yaml:"listen_addr"`
	MetricsAddr       string  `yaml:"metrics_addr"`
	Threshold         float64 `yaml:"threshold"`
	UpstreamProxy     string  `yaml:"upstream_proxy"`
	PrayerCity        string  `yaml:"prayer_city"`
	PrayerCountry     string  `yaml:"prayer_country"`
	PrayerTimezone    string  `yaml:"prayer_timezone"`
	StrictMode        bool    `yaml:"strict_mode"`
	StrictDelta       float64 `yaml:"strict_delta"`
	PreWindowMinutes  int     `yaml:"pre_window_minutes"`
	PostWindowMinutes int     `yaml:"post_window_minutes"`
	Blocklist         struct {
		SourceURL    string `yaml:"source_url"`
		SigningKey   string `yaml:"signing_key"`
		CachePath    string `yaml:"cache_path"`
		MetadataPath string `yaml:"metadata_path"`
	} `yaml:"blocklist"`
}

type Proxy struct {
	Config        Config
	TextScorer    ml.TextScorer
	ImageScorer   ml.ImageScorer
	Blocklist     blocklist.Blocklist
	PrayerManager prayertimes.Manager
	BlockedCount  atomic.Uint64
	AllowedCount  atomic.Uint64
	client        *http.Client
}

type Decision struct {
	Blocked    bool    `json:"blocked"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
	Threshold  float64 `json:"threshold"`
}

func LoadConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:8080"
	}
	if cfg.MetricsAddr == "" {
		cfg.MetricsAddr = "127.0.0.1:9090"
	}
	if cfg.Threshold == 0 {
		cfg.Threshold = 0.70
	}
	if cfg.StrictDelta == 0 {
		cfg.StrictDelta = 0.15
	}
	if cfg.Blocklist.CachePath == "" {
		cfg.Blocklist.CachePath = filepath.Join(filepath.Dir(path), "blocklist.json")
	}
	if cfg.Blocklist.MetadataPath == "" {
		cfg.Blocklist.MetadataPath = filepath.Join(filepath.Dir(path), "blocklist.meta.json")
	}
	return cfg, nil
}

func NewProxy(cfg Config) (*Proxy, error) {
	proxy := &Proxy{
		Config:      cfg,
		TextScorer:  ml.NewTextScorer(),
		ImageScorer: ml.NewImageScorer(),
		PrayerManager: prayertimes.Manager{
			PreWindowMinutes:  cfg.PreWindowMinutes,
			PostWindowMinutes: cfg.PostWindowMinutes,
			StrictDelta:       cfg.StrictDelta,
		},
		client: &http.Client{Timeout: 20 * time.Second},
	}
	if cfg.UpstreamProxy != "" {
		proxyURL, err := url.Parse(cfg.UpstreamProxy)
		if err != nil {
			return nil, err
		}
		proxy.client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}
	if cfg.Blocklist.CachePath != "" {
		list, err := blocklist.NewSyncClient(cfg.Blocklist.SourceURL, cfg.Blocklist.SigningKey, cfg.Blocklist.CachePath, cfg.Blocklist.MetadataPath, nil).LoadCached()
		if err == nil {
			proxy.Blocklist = list
		}
	}
	return proxy, nil
}

func (p *Proxy) StartWithReload(configPath string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleHTTP)
	server := &http.Server{Addr: p.Config.ListenAddr, Handler: mux}
	metricsServer := &http.Server{Addr: p.Config.MetricsAddr, Handler: http.HandlerFunc(p.handleMetrics)}

	errCh := make(chan error, 2)
	go func() { errCh <- server.ListenAndServe() }()
	go func() { errCh <- metricsServer.ListenAndServe() }()
	go p.reloadLoop(configPath)
	log.Printf("HalalBrowse proxy listening on %s (metrics %s)", p.Config.ListenAddr, p.Config.MetricsAddr)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		metricsServer.Shutdown(ctx)
	}()

	for {
		err := <-errCh
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			continue
		}
		return err
	}
}

func (p *Proxy) reloadLoop(configPath string) {
	sigCh := make(chan os.Signal, 1)
	signalNotify(sigCh, syscall.SIGHUP)
	for range sigCh {
		cfg, err := LoadConfig(configPath)
		if err != nil {
			log.Printf("config reload failed: %v", err)
			continue
		}
		updated, err := NewProxy(cfg)
		if err != nil {
			log.Printf("proxy rebuild failed: %v", err)
			continue
		}
		p.Config = updated.Config
		p.TextScorer = updated.TextScorer
		p.ImageScorer = updated.ImageScorer
		p.Blocklist = updated.Blocklist
		p.PrayerManager = updated.PrayerManager
		p.client = updated.client
		log.Printf("reloaded config from %s", configPath)
	}
}

func signalNotify(ch chan<- os.Signal, sig ...os.Signal) {
	signal.Notify(ch, sig...)
}

func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
		return
	}
	if p.Blocklist.Matches(r.URL.String()) {
		p.BlockedCount.Add(1)
		writeBlockedResponse(w, Decision{Blocked: true, Reason: "matched blocklist", Threshold: p.Config.Threshold})
		return
	}
	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	if !outReq.URL.IsAbs() {
		outReq.URL.Scheme = "http"
		outReq.URL.Host = r.Host
	}
	resp, err := p.client.Do(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	decision := p.scoreResponse(w, r, body, p.strictModeActive())
	if decision.Blocked {
		p.BlockedCount.Add(1)
		writeBlockedResponse(w, decision)
		return
	}
	p.AllowedCount.Add(1)
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	if p.Blocklist.Matches("https://" + r.Host) {
		p.BlockedCount.Add(1)
		writeBlockedResponse(w, Decision{Blocked: true, Reason: "matched blocklist host", Threshold: p.Config.Threshold})
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		return
	}
	defer targetConn.Close()
	fmt.Fprint(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	go io.Copy(targetConn, clientConn)
	io.Copy(clientConn, targetConn)
}

func (p *Proxy) strictModeActive() bool {
	if !p.Config.StrictMode {
		return false
	}
	schedule, err := prayertimes.LoadFallbackCSV("date,city,timezone,fajr,dhuhr,asr,maghrib,isha\n2026-04-22,Toulouse,UTC,05:30,12:00,15:30,18:45,20:15\n", time.Now().UTC(), p.Config.PrayerCity)
	if err != nil {
		return false
	}
	strict, _ := p.PrayerManager.StrictModeAt(schedule, time.Now().UTC())
	return strict
}

func (p *Proxy) scoreResponse(_ http.ResponseWriter, _ *http.Request, body []byte, strict bool) Decision {
	score, _ := scoreBody(p.TextScorer, body)
	threshold := p.PrayerManager.EffectiveThreshold(p.Config.Threshold, strict)
	if score.Confidence < threshold {
		return Decision{Blocked: true, Reason: strings.Join(score.Reasons, "; "), Confidence: score.Confidence, Threshold: threshold}
	}
	return Decision{Blocked: false, Reason: strings.Join(score.Reasons, "; "), Confidence: score.Confidence, Threshold: threshold}
}

func (p *Proxy) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "# HELP halalbrowse_blocked_total Total blocked responses\n")
	fmt.Fprintf(w, "# TYPE halalbrowse_blocked_total counter\n")
	fmt.Fprintf(w, "halalbrowse_blocked_total %d\n", p.BlockedCount.Load())
	fmt.Fprintf(w, "# HELP halalbrowse_allowed_total Total allowed responses\n")
	fmt.Fprintf(w, "# TYPE halalbrowse_allowed_total counter\n")
	fmt.Fprintf(w, "halalbrowse_allowed_total %d\n", p.AllowedCount.Load())
}

func writeBlockedResponse(w http.ResponseWriter, decision Decision) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, "<html><body><h1>HalalBrowse blocked this page</h1><p>%s</p><p>confidence=%.2f threshold=%.2f</p></body></html>", decision.Reason, decision.Confidence, decision.Threshold)
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func parseRequestLine(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func readCONNECTTarget(conn net.Conn) string {
	reader := bufio.NewReader(conn)
	line, _ := reader.ReadString('\n')
	return parseRequestLine(line)
}
