# Loom walkthrough — target: 3 minutes 30 seconds

## 0:00–0:30 — Product

“Hi Tommy, I’m Olamitunde. This is MarketPulse AI, a service that turns unstructured financial news into structured signals: summary, sentiment, confidence, ticker, and events such as earnings or guidance changes. I chose this because it mirrors the core engineering problem at Benzinga: extracting useful insight from time-sensitive financial text and delivering it at low latency.”

Show: `README.md`, then the architecture diagram.

## 0:30–1:25 — Architecture and code

“The system has two independently deployable services. Go owns the public hot path because it is efficient under high concurrency. In `handler.go` I validate and cap payloads, use content-addressed caching, propagate request context, enforce timeouts, and expose metrics. Python owns inference so the ML team can experiment quickly. Its HTTP contract is model-agnostic: today it uses an explainable financial baseline; tomorrow I can promote FinBERT or a fine-tuned event classifier without changing clients.”

Show: `internal/api/handler.go` lines around `analyze`, then `services/ml/app.py` around `analyze`.

## 1:25–2:00 — Live demo

Run `make demo`. Point out the structured response and `X-Cache: MISS`. Run it again and point out `X-Cache: HIT`. Optionally open `http://localhost:18080/metrics`. The host uses port 18080 to avoid the local Jenkins service on port 8080; the Go container still listens on 8080.

“The cache key is a SHA-256 digest of ticker and content, so syndicated duplicates do not repeatedly consume inference capacity. The second identical request bypasses Python.”

## 2:00–2:45 — Biggest challenge

“The hardest design problem was balancing model iteration with predictable API latency. Putting model logic directly in the Go service would couple deployments and make Python ML tooling awkward. A single Python app, on the other hand, would make the public delivery layer compete with inference for resources. I separated the services behind a strict schema and added deadlines, a 502 failure boundary, and cache-aside behavior. That lets each tier scale independently and makes failure explicit.”

Show: `internal/client/ml.go` timeout/context and `docker-compose.yml`.

## 2:45–3:20 — Go experience

“In Go I’m comfortable building HTTP microservices using the standard library, goroutine-safe state, atomic operational counters, context propagation, and defensive server timeouts. Here, the TTL cache uses an RWMutex, counters use atomics, and the HTTP client is bounded. In production I’d replace the local cache with Redis and put Kafka ahead of inference for replay, ordering, and backpressure.”

Show: `internal/cache/cache.go`, briefly `cmd/gateway/main.go`.

## 3:20–3:35 — Close

“The repository also includes container builds, health checks, CI, and tests. My next step would be shadow-evaluating FinBERT against labeled Benzinga data and tracking per-event precision, drift, and p95 latency before promotion. Thanks for watching.”

## Recording checklist

- Close notifications and unrelated tabs; use 125–150% editor zoom.
- Start the containers and make the first demo request before recording to verify output, then restart to restore `MISS`.
- Keep the README, handler, Python model, cache, and terminal in five prepared tabs.
- Record at 1080p; keep the camera bubble away from the lower-right terminal output.
- Never claim production traffic or experience you cannot substantiate. Describe this as a production-minded prototype unless it has actually served users.
