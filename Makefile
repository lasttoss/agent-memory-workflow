# The diagram sources are HTML and Mermaid; a PNG is a build artifact.
.PHONY: test check diagram

test:
	go test -race -count=1 ./...

check:
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test -race -count=1 ./...

diagram:
	@if command -v chromium >/dev/null 2>&1; then B=chromium; elif command -v google-chrome >/dev/null 2>&1; then B=google-chrome; else echo "no chromium on PATH: open docs/diagrams/*.html in a browser"; exit 0; fi; \
	for f in docs/diagrams/*.html; do $$B --headless --screenshot="$${f%.html}.png" --window-size=1200,1000 "$$f" && echo "wrote $${f%.html}.png"; done
