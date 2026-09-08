GATEWAY_URL ?= http://localhost:18080

.PHONY: test up down demo

test:
	cd services/ml && python -m unittest -v

up:
	docker compose up --build -d

down:
	docker compose down

demo:
	@curl -s -D /tmp/marketpulse-headers -X POST $(GATEWAY_URL)/v1/analyze \
		-H 'Content-Type: application/json' --data-binary @data/lilly-earnings.json | python -m json.tool
	@grep -i '^x-cache:' /tmp/marketpulse-headers
