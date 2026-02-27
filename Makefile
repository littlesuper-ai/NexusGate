.PHONY: dev-server dev-web test test-server test-web lint build clean docker-up docker-down

# ─── Development ──────────────────────────────────────────────

dev-server:
	cd server && go run ./cmd/nexusgate

dev-web:
	cd web && npm run dev

# ─── Testing ──────────────────────────────────────────────────

test: test-server test-web

test-server:
	cd server && go test ./... -v -race

test-web:
	cd web && npm test

test-web-watch:
	cd web && npm run test:watch

# ─── Lint & Check ────────────────────────────────────────────

lint: lint-server lint-web

lint-server:
	cd server && go vet ./...

lint-web:
	cd web && npx vue-tsc --noEmit

# ─── Build ────────────────────────────────────────────────────

build: build-server build-web

build-server:
	cd server && CGO_ENABLED=0 go build -ldflags="-s -w" -o nexusgate ./cmd/nexusgate

build-web:
	cd web && npm run build

# ─── Docker ───────────────────────────────────────────────────

docker-up:
	cd deploy && docker compose up -d

docker-down:
	cd deploy && docker compose down

docker-logs:
	cd deploy && docker compose logs -f

docker-build:
	cd deploy && docker compose build

# ─── Cleanup ──────────────────────────────────────────────────

clean:
	rm -f server/nexusgate
	rm -rf web/dist web/node_modules/.vite
