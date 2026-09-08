# MarketPulse AI — Five-Minute Presentation and Defense Guide

This guide prepares the presenter to explain MarketPulse AI clearly and defend its engineering decisions. Do not memorize every word. Understand the ideas, rehearse twice, and speak naturally.

> **Important:** Describe this as a production-minded prototype built for the Benzinga exercise. Do not claim that it served real users, processed production traffic, or used proprietary Benzinga data. Only state professional experience that is true and can be supported with examples.

## The one-sentence explanation

> MarketPulse AI converts unstructured financial news into structured, API-ready signals through a reliable Go gateway and an independently deployable Python inference service.

## Before recording

Prepare these five tabs in this order:

1. `README.md` — architecture diagram.
2. `internal/api/handler.go` — the `analyze` function.
3. `services/ml/app.py` — the `analyze` and `has_required_text` functions.
4. `internal/cache/cache.go` — concurrency-safe TTL cache.
5. Terminal — enlarged and ready to run the demo.

Reset the services so the first recorded request is a cache miss:

```bash
# Run from the cloned repository directory.
docker compose down
docker compose up --build -d
```

Confirm it is ready with `docker compose ps`, but do not run `make demo` until recording.

## Five-minute walkthrough

### 0:00–0:35 — Introduction and product purpose

Show the README title and architecture.

Say:

> Hi Tommy, I’m Aderonke Joy Ogundipe. This is MarketPulse AI, a production-minded financial-news intelligence prototype. It accepts an unstructured market article and returns a structured signal containing the ticker, sentiment, confidence, detected events, a summary, and model version. I chose this problem because it closely reflects Benzinga’s work: turning time-sensitive financial text into reliable data that downstream products and APIs can consume.

Key idea to understand: the product is not merely a sentiment demo. Its value is the complete path from text inference to a defensive, observable API.

### 0:35–1:25 — Architecture and main design decision

Point through the architecture diagram from left to right.

Say:

> I separated the system into two independently deployable services. Go owns the public, high-concurrency delivery path: request validation, payload limits, caching, deadlines, error handling, and metrics. Python owns inference because Python has the strongest ML ecosystem and allows rapid model experimentation. The HTTP boundary is deliberately model-agnostic, so I can replace the explainable baseline with FinBERT, a fine-tuned classifier, or an LLM without changing public clients.
>
> The tradeoff is one additional network hop and more operational components. I accepted that cost because the services can scale independently, model failures are isolated, and backend releases do not need to be coupled to every model experiment.

Why this answer is strong: it states the decision, benefit, and cost. Do not describe a choice without acknowledging its tradeoff.

### 1:25–2:15 — Go gateway walkthrough

Open `internal/api/handler.go` and show the `analyze` method.

Say:

> The Go handler caps request bodies at one megabyte, disallows unknown JSON fields, and rejects whitespace-only content. It computes a SHA-256 content key before inference. A cache hit returns immediately; a miss calls Python using the incoming request context. The ML client has a three-second timeout, and upstream errors become a sanitized 502 rather than leaking internal details. The gateway also tracks total requests, cache hits, and recent inference latency.

Briefly point to:

- `http.MaxBytesReader` — resource protection;
- `DisallowUnknownFields` — strict contract;
- `strings.TrimSpace` — meaningful validation;
- `contentKey` and cache lookup — duplicate-work control;
- `r.Context()` — cancellation propagation;
- `X-Cache` — visible demo behavior.

Do not read the code line by line. Explain why each group of lines exists.

### 2:15–2:55 — Python inference walkthrough

Open `services/ml/app.py` and show `analyze`.

Say:

> The current inference implementation is a deterministic financial baseline. It uses weighted domain vocabulary, bounds the score with `tanh`, detects events with explicit patterns, and produces an extractive summary. I chose this baseline because it is fast, explainable, reproducible, and requires no external credentials for the demo. I am not presenting it as state-of-the-art ML. The architecture is designed so stronger models can be evaluated and promoted behind the same contract.
>
> The Python service repeats important validation and applies a five-second socket timeout, so it remains defensive even when called directly rather than through Go.

If asked why not use an LLM immediately: the baseline establishes deterministic behavior, latency, and evaluation plumbing before introducing model cost and nondeterminism.

### 2:55–3:40 — Live demonstration

Open the terminal and run:

```bash
make demo
make demo
curl http://localhost:18080/metrics
```

Say while showing the first response:

> This Eli Lilly article is classified as positive and tagged with earnings and guidance-change events. The response also includes confidence and model version for traceability. The first request reports `X-Cache: MISS`, so Python performed inference.

After the second command:

> The identical request now reports `X-Cache: HIT`. The SHA-256 content key prevents duplicate inference when the same story is requested or syndicated repeatedly.

While showing metrics:

> The metrics confirm two requests and one cache hit. In a production system I would use Prometheus histograms for p50, p95, and p99 latency rather than only the latest value.

### 3:40–4:20 — Biggest technical challenge

Show the architecture again or `internal/client/ml.go`.

Say:

> The main technical challenge was balancing rapid model iteration with predictable API reliability. Keeping everything in Python would make public traffic and inference compete in the same runtime. Embedding model behavior directly in Go would reduce access to Python’s ML tooling and couple model and API releases. I solved this with a strict service boundary, bounded calls, cache-aside behavior, and explicit failure mapping. That gives each tier an independent scaling and deployment path.

This is the most important section. Deliver it calmly: **problem → alternatives → decision → tradeoff → outcome**.

### 4:20–4:45 — Go experience

Use exactly one version below and customize it truthfully before recording.

If you have professional Go experience:

> I have used Go in [REAL PROJECT OR COMPANY] to build [REAL SERVICE OR TOOL]. My responsibilities included [TWO TRUE RESPONSIBILITIES], and the main operational consideration was [REAL CHALLENGE]. In this project I applied the same principles: standard-library HTTP services, context propagation, defensive timeouts, mutex-protected shared state, atomic counters, and small testable packages.

If your Go experience is mainly projects and hands-on learning:

> My strongest production background is [YOUR TRUE PRIMARY LANGUAGE OR AREA], while my Go experience comes from building services and infrastructure-focused projects such as this one. I’m comfortable with Go’s standard HTTP library, context propagation, concurrency-safe state, table-driven testing, modules, and container builds. I’m deliberately not overstating the length of that experience; what I can demonstrate is that I understand how to design, test, and operate a Go service, and I can become productive quickly.

Never say “I have X years of production Go” unless it is accurate and supported by your résumé.

### 4:45–5:00 — Production roadmap and close

Say:

> My next step would be to create a labeled financial-news evaluation set, shadow-test FinBERT and a fine-tuned event classifier, and compare precision, recall, calibration, and latency before promotion. For scale, I would add Kafka, Redis, OpenSearch, and separate EKS scaling policies for the Go and inference tiers. Thank you for watching—I’d be happy to discuss the tradeoffs in more detail.

## Likely technical questions and strong answers

### “Why Go instead of only Python?”

> Python is excellent for model development, but Go gives the public API a small memory footprint, straightforward concurrency, a compiled deployment artifact, and strong standard-library networking. The split also prevents model experimentation from forcing public API changes. For a low-volume prototype one Python service would be simpler; this design demonstrates the intended high-throughput production direction.

### “Is this really machine learning?”

> The current implementation is an explainable NLP baseline, not a trained neural model. That is intentional for a reproducible, credential-free demonstration. In a real ML lifecycle I would use it as a benchmark, build a labeled dataset, compare FinBERT or a fine-tuned classifier against it, calibrate confidence, and promote only when offline and shadow metrics justify the added complexity.

### “How would you evaluate it?”

> I would split data chronologically to reduce leakage, measure precision, recall, and F1 per event and sentiment class, inspect confusion matrices, evaluate probability calibration, and create slices for tickers, publishers, article types, and market regimes. I would also track business-oriented measures such as alert usefulness and false-positive cost. In production I would monitor input drift, confidence distributions, latency, and delayed-label performance.

### “Why use a SHA-256 cache key?”

> It provides a stable fixed-size key derived from the actual inference inputs and makes exact duplicate detection inexpensive. The ticker is included because it affects interpretation. In production I would also normalize publisher-specific formatting and use article IDs for idempotency. SHA-256 only solves exact duplication; semantic near-duplicates require embeddings or locality-sensitive methods.

### “What happens if Python fails?”

> The Go client stops waiting after three seconds, respects caller cancellation, and returns a sanitized 502. Failed results are not cached. At production scale I would add bounded retries only for safe transient failures, circuit breaking, queue-based asynchronous processing where appropriate, and alerts tied to error rate and latency objectives.

### “Why an in-memory cache?”

> It keeps the prototype dependency-free and demonstrates cache-aside behavior. It is process-local, disappears on restart, and cannot coordinate replicas. Redis would be the production replacement, with TTL policy, capacity limits, hit-rate monitoring, and protection against cache stampedes.

### “How would this scale to millions of requests?”

> I would separate synchronous reads from asynchronous ingestion. Kafka would buffer and replay articles; inference workers would autoscale from queue depth and latency; results would be persisted and cached in Redis; Go API replicas would remain stateless behind a load balancer. CPU and GPU inference would use different EKS node groups. Load tests would establish saturation points and drive resource and autoscaling settings.

### “How would you add RAG?”

> I would chunk and embed trusted financial documents, store vectors and metadata in OpenSearch kNN, retrieve with ticker and time filters, rerank results, and require generated summaries to cite retrieved sources. I would evaluate retrieval recall separately from answer faithfulness and prevent future-dated information from leaking into historical evaluation.

### “What are the current weaknesses?”

> The lexical model has limited handling of negation, nuance, and entity relationships. The cache is instance-local, metrics are minimal, and there is no persistence, authentication, streaming, or rate limiting. Those limitations are documented. The prototype demonstrates architectural boundaries and operational thinking, while the roadmap explains how I would evolve it based on measured requirements.

### “Why does FDA approval produce both a regulatory event and positive sentiment?”

> Event type and sentiment represent different dimensions. FDA approval is a regulatory event and is often positive news, so both outputs can be correct. The weakness is that the baseline applies a fixed positive weight regardless of context. A trained contextual model and labeled regulatory examples would handle exceptions more reliably.

### “What tests protect the system?”

> Python has inference and whitespace-validation tests. Go tests cache expiry, API validation, cache miss-to-hit behavior, metrics, sanitized 502 mapping, successful ML-response decoding, timeouts, and non-200 upstream responses. The Docker build enforces formatting, runs Go tests, and compiles the binary; GitHub Actions repeats formatting, Go, Python, and Compose validation.

## What the presenter must understand

Before recording, be able to explain these terms without reading:

- **Cache-aside:** the application checks cache first, computes on a miss, then stores the result.
- **Context propagation:** cancellation or deadlines from the incoming request are carried into the downstream request.
- **Backpressure:** controlling incoming work so a slow dependency cannot exhaust the service.
- **Idempotency:** processing the same logical event repeatedly without creating different side effects.
- **Model drift:** production input or outcome patterns changing from the data used to evaluate the model.
- **Shadow deployment:** running a candidate model on live inputs without letting its output affect users.
- **p95 latency:** the response time at or below which 95% of requests complete.
- **Confidence calibration:** ensuring a stated probability corresponds to how often the prediction is correct.

## Final delivery advice

- Aim for 4:20–4:45, leaving margin below Loom’s five-minute limit.
- Speak to the architecture and decisions; do not narrate every code line.
- Keep the editor zoomed to 125–150% and notifications disabled.
- Run the commands exactly as shown; port `18080` avoids the local Jenkins collision on `8080`.
- If something fails during recording, explain how you would diagnose it instead of becoming silent.
- Do not use answers you do not understand. Rehearse each follow-up answer in your own words.
- Make sure the presenter’s name, résumé, claimed experience, repository explanation, and submission identity are consistent and truthful.
