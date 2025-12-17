.PHONY: dev-ui build-ui start-ui add-indexes

dev-ui:
	cd app && pnpm dev

build-ui:
	cd app && pnpm build

start-ui:
	cd app && pnpm start

add-indexes:
	go run ./cmd/add-indexes
