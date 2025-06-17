.PHONY: build dev deploy clean assets

# Build the WASM binary and generate assets
build: clean assets
	tinygo build -o build/app.wasm -target wasm -no-debug main.go

# Generate JavaScript assets for Cloudflare Workers
assets:
	mkdir -p build
	workers-assets-gen

# Run development server
dev: build
	wrangler dev

# Deploy to Cloudflare
deploy: build
	wrangler deploy

# Clean build artifacts
clean:
	rm -rf build

# Install workers-assets-gen tool
install-tools:
	go install github.com/syumai/workers/cmd/workers-assets-gen@latest

# Database commands
db-create:
	wrangler d1 create openlist-db

db-migrate-local:
	wrangler d1 migrations apply openlist-db --local

db-migrate-remote:
	wrangler d1 migrations apply openlist-db --remote

db-query-local:
	wrangler d1 execute openlist-db --local --command "SELECT * FROM users LIMIT 10;"

db-query-remote:
	wrangler d1 execute openlist-db --remote --command "SELECT * FROM users LIMIT 10;" 