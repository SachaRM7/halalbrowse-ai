# HalalBrowse AI

HalalBrowse AI is a privacy-first prototype implementing the Phase 1 core engine and a lightweight cross-platform shell for the broader PRD. This build includes:

- Go core packages for text/image scoring, blocklist verification, and prayer-time strict mode
- A working Go HTTP proxy/CLI with metrics, blocklist pull, config reload via SIGHUP, and threshold-based blocking
- WebExtension source files for Chrome/Firefox popup + content blocking flow
- A React Native Android screen showing the intended local VPN/filter status UX
- Basic Go unit tests for blocklist sync, prayer strict mode, and proxy scoring

Project layout:

- core/ml: deterministic offline scorers that emulate bundled on-device model behavior
- core/blocklist: signed JSON blocklist parsing and cached sync handling
- core/prayertimes: prayer schedule strict-mode manager with Aladhan fetch + CSV fallback loader
- cli: local HTTP proxy and CLI entrypoints
- extension: Manifest V3 browser extension source
- android: React Native screen for Android MVP shell

Quick start:

1. Install Go 1.22+
2. Copy cli/config.yaml.example to ~/.halalbrowse/config.yaml
3. Edit the blocklist source URL and signing key
4. Pull the latest blocklist:
   go run ./cli blocklist pull --config ~/.halalbrowse/config.yaml
5. Start the proxy:
   go run ./cli serve --config ~/.halalbrowse/config.yaml
6. Point your browser or terminal tools at http://127.0.0.1:8080
7. View Prometheus metrics at http://127.0.0.1:9090

Run tests:

- go test ./...

Notes:

- The ML layer is implemented as deterministic local heuristics so the code runs without external model binaries; the interfaces are ready to be swapped with ONNX/TFLite-backed inference.
- HTTPS CONNECT is supported for host-level blocklist enforcement; deep TLS content inspection would require certificate management beyond this MVP.
- The browser extension and Android app are source-only MVP shells intended to sit on top of the shared core protocol.
