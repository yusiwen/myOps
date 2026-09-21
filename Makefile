.PHONY: build dev tailwind lint fmt clean

BINARY=myops

build: tailwind
	go build -o bin/$(BINARY) ./cmd/myops

dev: tailwind
	air

tailwind:
	tailwindcss -i web/static/css/input.css -o web/static/css/tailwind.css

lint:
	golangci-lint run ./...

fmt:
	nix fmt

clean:
	rm -rf bin/ tmp/ web/static/css/tailwind.css
