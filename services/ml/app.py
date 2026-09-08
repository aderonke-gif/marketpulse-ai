"""Dependency-free financial-news inference service.

The baseline is deliberately explainable and deterministic. The HTTP contract is model-
agnostic, so a FinBERT/LLM implementation can be promoted behind it without touching Go.
"""

from __future__ import annotations

import json
import math
import re
import time
from collections import Counter
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

POSITIVE = {
    "beat": 1.6, "beats": 1.6, "growth": 1.1, "record": 1.0,
    "raises": 1.4, "raised": 1.4, "surge": 1.3, "profit": 1.0,
    "approval": 1.6, "approved": 1.6, "outperform": 1.3, "strong": 0.8,
}
NEGATIVE = {
    "miss": -1.6, "misses": -1.6, "cuts": -1.4, "cut": -1.4,
    "decline": -1.1, "loss": -1.1, "investigation": -1.5,
    "downgrade": -1.3, "layoffs": -1.0, "weak": -0.8, "risk": -0.6,
}
EVENT_PATTERNS = {
    "earnings": r"\b(earnings|eps|revenue|quarter|q[1-4])\b",
    "guidance_change": r"\b(raises?|raised|lowers?|lowered|cuts?|increases?|increased)\b.{0,60}\b(guidance|outlook|forecast)\b|\b(guidance|outlook|forecast)\b.{0,60}\b(raises?|raised|lowers?|lowered|cuts?|increases?|increased)\b",
    "merger_acquisition": r"\b(acquire|acquisition|merger|buyout)\b",
    "regulatory": r"\b(fda|sec|regulator|approval|investigation)\b",
    "analyst_rating": r"\b(upgrade|downgrade|price target|outperform)\b",
}
WORD_RE = re.compile(r"[a-z0-9$%]+")


def analyze(article: dict[str, Any]) -> dict[str, Any]:
    headline = str(article.get("headline", "")).strip()
    body = str(article.get("body", "")).strip()
    ticker = str(article.get("ticker", "")).upper()
    text = f"{headline}. {body}".lower()
    counts = Counter(WORD_RE.findall(text))

    raw = sum(counts[word] * weight for word, weight in POSITIVE.items())
    raw += sum(counts[word] * weight for word, weight in NEGATIVE.items())
    # Saturation keeps repeated terms from producing unjustified certainty.
    score = math.tanh(raw / 3.5)
    sentiment = "positive" if score > 0.2 else "negative" if score < -0.2 else "neutral"
    events = [name for name, pattern in EVENT_PATTERNS.items() if re.search(pattern, text)]
    confidence = min(0.97, 0.55 + abs(score) * 0.3 + min(len(events), 2) * 0.06)

    first_sentence = re.split(r"(?<=[.!?])\s+", body)[0] if body else headline
    summary = headline.rstrip(".") + "."
    if first_sentence and first_sentence.lower().rstrip(".") != headline.lower().rstrip("."):
        summary += " " + first_sentence[:220].rstrip(".") + "."

    return {
        "article_id": str(article.get("id", "")),
        "ticker": ticker,
        "sentiment": sentiment,
        "score": round(score, 4),
        "events": events,
        "summary": summary,
        "confidence": round(confidence, 4),
        "model": "financial-lexicon-baseline-v1",
    }


class Handler(BaseHTTPRequestHandler):
    server_version = "MarketPulseML/1.0"

    def do_GET(self) -> None:
        if self.path == "/health":
            self.respond(200, {"status": "ok", "service": "ml-inference"})
        else:
            self.respond(404, {"error": "not found"})

    def do_POST(self) -> None:
        if self.path != "/analyze":
            self.respond(404, {"error": "not found"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length <= 0 or length > 1_048_576:
                raise ValueError("body must be between 1 byte and 1 MB")
            request = json.loads(self.rfile.read(length))
            article = request.get("article", {})
            if not article.get("headline") or not article.get("body"):
                self.respond(422, {"error": "headline and body are required"})
                return
            started = time.perf_counter()
            result = analyze(article)
            result["inference_ms"] = round((time.perf_counter() - started) * 1000, 3)
            self.respond(200, result)
        except (ValueError, json.JSONDecodeError) as exc:
            self.respond(400, {"error": f"invalid request: {exc}"})

    def respond(self, status: int, payload: dict[str, Any]) -> None:
        encoded = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(encoded)))
        self.end_headers()
        self.wfile.write(encoded)

    def log_message(self, fmt: str, *args: Any) -> None:
        print(json.dumps({"message": fmt % args, "service": "ml-inference"}))


if __name__ == "__main__":
    print(json.dumps({"message": "ML service listening", "port": 8000}))
    ThreadingHTTPServer(("0.0.0.0", 8000), Handler).serve_forever()
