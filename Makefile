.PHONY: dev build web-dev api-dev

dev:
	pnpm lint && pnpm dev

build:
	pnpm --filter relay-agent-workspace build
	cd apps/api && go build -o main .

web-dev:
	pnpm --filter relay-agent-workspace dev

api-dev:
	cd apps/api && go run main.go
