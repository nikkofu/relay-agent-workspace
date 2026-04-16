.PHONY: dev build web-dev api-dev

dev:
	pnpm dev

build:
	pnpm --filter web build
	cd apps/api && go build -o main .

web-dev:
	pnpm --filter web dev

api-dev:
	cd apps/api && go run main.go
