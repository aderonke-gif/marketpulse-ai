# MarketPulse AI

MarketPulse AI is a production-minded financial-news intelligence prototype built for a Benzinga AI/Machine Learning Engineer hiring exercise. It converts unstructured market news into low-latency, structured signals: sentiment, confidence, ticker, event categories, and a concise summary.

> **Status:** Working interview prototype—not a deployed Benzinga product and not a claim of prior production traffic.

## Why this repository exists

Benzinga requested a **2–5 minute screen-sharing Loom video** that covers:

1. The product and its purpose.
2. The codebase, architecture, and key technical components.
3. The biggest technical challenge and its solution.
4. Experience with Go and relevant use cases.

The target role combines financial NLP, LLM/RAG systems, production model serving, Python and Go microservices, real-time pipelines, AWS, Kubernetes, observability, and end-to-end ownership. This project was scoped to demonstrate those themes in a system that can be explained and run live within the short video.

The prepared narration is in [`docs/LOOM_SCRIPT.md`](docs/LOOM_SCRIPT.md), and final checks are in [`docs/SUBMISSION_CHECKLIST.md`](docs/SUBMISSION_CHECKLIST.md).

## Product overview

Financial news arrives as human-readable text, while trading, alerting, and research systems need predictable structured data. MarketPulse AI accepts an article and returns:

- normalized ticker and article identity;
- positive, neutral, or negative sentiment with a bounded score;
- detected events such as earnings, guidance changes, M&A, regulatory activity, and analyst ratings;
- a concise summary;
- confidence and model-version metadata for traceability.

The supplied demo turns an Eli Lilly earnings article into a positive signal tagged as `earnings` and `guidance_change`, ready for an API consumer or data pipeline.

## Architecture

```mermaid
flowchart LR
    A[News feed / API client] -->|POST /v1/analyze| B[Go API gateway]
    B --> C{SHA-256 TTL cache}
    C -->|hit| B
    C -->|miss| D[Python NLP service]
    D --> E[Sentiment + events + summary]
    E --> C
    B --> F[Structured JSON signal]
    B --> G[/health + /metrics]
```

### The central technical decision

The main challenge is balancing rapid model iteration with predictable API latency and reliability.

- **Go owns the public hot path:** concurrent HTTP traffic, strict JSON validation, 1 MB request limits, context propagation, upstream timeouts, cache access, logs, health checks, and metrics.
- **Python owns inference:** it keeps experimentation close to the ML ecosystem behind a stable, model-agnostic HTTP contract.
- **The boundary isolates change and failure:** FinBERT, a fine-tuned transformer, or an LLM can replace the baseline without changing the public API. Go limits inference calls to three seconds and maps upstream failures to a controlled `502`.
- **Content-addressed caching controls duplicate work:** a SHA-256 key derived from ticker, headline, and body prevents repeated inference for identical syndicated stories.

The services can be tested, deployed, and scaled independently. The tradeoff is an extra network hop and more operational components.

## Request lifecycle

1. A client sends an article to `POST /v1/analyze` on port `8080`.
2. Go rejects malformed, oversized, or incomplete input.
3. It hashes the content and checks the concurrency-safe TTL cache.
4. A cache miss invokes Python on port `8000` with a three-second deadline.
5. Python calculates sentiment, extracts events, produces a summary, and reports its model version.
6. Go stores and returns the result. The same request then returns `X-Cache: HIT` without invoking Python.
7. `/metrics` exposes request count, cache hits, and recent inference latency in Prometheus format.

## Repository map

```text
.
├── cmd/gateway/main.go          # Go entry point and server deadlines
├── internal/api/handler.go      # Validation, cache-aside flow, metrics
├── internal/cache/cache.go      # Goroutine-safe TTL cache
├── internal/client/ml.go        # Context-aware inference client
├── internal/model/model.go      # Go request/response schemas
├── services/ml/app.py           # NLP inference and Python HTTP service
├── services/ml/test_app.py      # Python inference tests
├── data/lilly-earnings.json     # Live-demo request
├── Dockerfile.gateway           # Tested multi-stage Go build
├── docker-compose.yml           # Two-service orchestration and health checks
├── .github/workflows/ci.yml     # Go, Python, and Compose CI
└── docs/                        # Loom script and submission checklist
```

## Model implementation

The current `financial-lexicon-baseline-v1` is deterministic and dependency-free. It uses a weighted financial vocabulary, bounded `tanh` sentiment scoring, regular-expression event detection, and extractive summarization. This keeps the demo fast, reproducible, inspectable, and free of external API keys.

It is a baseline, not a claim that lexicon analysis is state-of-the-art NLP. The service boundary allows a stronger model to be shadow-tested and promoted without disrupting clients.

## Run the complete demo

Requirements: Docker Desktop with Docker Compose.

```bash
git clone https://github.com/aderonke-gif/marketpulse-ai.git
cd marketpulse-ai
docker compose up --build -d
make demo
make demo
curl http://localhost:8080/metrics
```

Expected behavior:

- the first `make demo` reports `X-Cache: MISS`;
- the second reports `X-Cache: HIT`;
- the response identifies positive LLY earnings and guidance-change signals;
- metrics report two requests, one cache hit, and inference latency.

Example response:

```json
{
  "article_id": "benzinga-demo-001",
  "ticker": "LLY",
  "sentiment": "positive",
  "score": 0.9336,
  "events": ["earnings", "guidance_change"],
  "summary": "Eli Lilly beats Q1 earnings estimates and raises full-year guidance. Revenue growth was strong after record demand for its diabetes and obesity portfolio.",
  "confidence": 0.9501,
  "model": "financial-lexicon-baseline-v1"
}
```

Stop with `docker compose down`.

## Testing and verification

```bash
make test                    # Python tests
go test ./...                # Go tests when Go 1.23+ is installed
docker compose build gateway # Go tests and compilation in Docker
```

Verified during initial development:

- four Python inference and validation tests passed;
- all Go packages compiled and seven cache, API, and client tests passed;
- both containers became healthy;
- the end-to-end request returned the documented signal;
- repeated inference changed from `MISS` to `HIT`;
- observed demo inference latency was 8 ms on the development machine.

## Production evolution

1. Ingest news through Kafka with idempotency, replay, backpressure, and a dead-letter queue.
2. Establish per-class precision, recall, F1, calibration, and slice metrics on labeled financial news.
3. Shadow-deploy FinBERT sentiment and a fine-tuned event classifier against the baseline.
4. Add OpenSearch kNN retrieval and grounded LLM summarization with citations.
5. Replace the local cache with Redis for horizontal scaling.
6. Track model version, drift, confidence distribution, errors, throughput, and p50/p95/p99 latency.
7. Deploy separate CPU/GPU workloads to EKS, store artifacts in S3, autoscale from queue depth, and support controlled promotion and rollback.

## Known limitations

- The baseline cannot understand nuanced context or entity relationships like a trained transformer.
- The cache is local to one gateway replica and is lost on restart.
- Metrics do not yet include histograms or per-model labels.
- Authentication, persistence, rate limiting, Kafka, and cloud infrastructure are outside this time-boxed prototype.
- The sample is synthetic; no proprietary Benzinga data is included.

## Guidance for AI agents and reviewers

When reviewing or extending this repository:

- Treat the hiring brief and Loom constraints above as the governing product context.
- Preserve the Go gateway/Python inference boundary unless measurements justify a change.
- Do not invent production traffic, accuracy, employment, or Go-experience claims.
- Keep the default demo runnable without paid APIs, credentials, or proprietary datasets.
- Add tests for behavior changes and verify the Python and containerized Go paths.
- Prioritize correctness, reliability, observability, measurable ML evaluation, and clear tradeoffs.
- Never commit secrets, personal data, model weights, or proprietary financial content.

## Ownership and attribution

The repository is hosted under the authorized `aderonke-gif` GitHub account. Commits for this hiring exercise are authored as **Olamitunde Temitope**. Repository ownership and commit authorship are intentionally distinct and should not be misrepresented.
