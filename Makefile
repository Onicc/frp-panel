.PHONY: web docs test security cross build controller agent clean

web:
	cd www && pnpm install --frozen-lockfile && pnpm build

docs:
	cd docs && pnpm install --frozen-lockfile && pnpm docs:build

test: web docs
	cd www && pnpm lint && pnpm typecheck && pnpm test && pnpm audit --prod --audit-level high
	go test -race ./...
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 ./...
	test -z "$$(gofmt -l -- .)"

security:
	go run github.com/securego/gosec/v2/cmd/gosec@v2.29.0 -quiet -exclude-generated -exclude=G104,G115,G204,G301,G302,G304,G122 ./...
	go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...

cross:
	@build_dir=$$(mktemp -d); trap 'rm -rf "$$build_dir"' EXIT; \
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do \
		suffix=; test "$${target%/*}" = windows && suffix=.exe || true; \
		GOOS=$${target%/*} GOARCH=$${target#*/} CGO_ENABLED=0 go build -trimpath -o "$$build_dir/agent-$${target%/*}-$${target#*/}$$suffix" ./cmd/frp-panel-agent; \
	done

build: web controller agent

controller:
	CGO_ENABLED=0 go build -trimpath -o dist/frp-panel ./cmd/frpp

agent:
	CGO_ENABLED=0 go build -trimpath -o dist/frp-panel-agent ./cmd/frp-panel-agent

clean:
	go clean -testcache
	rm -rf dist www/node_modules www/playwright-report www/test-results docs/node_modules docs/.vitepress/dist
