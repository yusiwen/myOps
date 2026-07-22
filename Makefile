.PHONY: build dev tailwind lint clean

BINARY=myops

build: tailwind
	go build -o bin/$(BINARY) ./cmd/myops

dev: tailwind
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air

tailwind:
	npx @tailwindcss/cli -i web/static/css/input.css -o web/static/css/tailwind.css

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ web/static/css/tailwind.css node_modules
