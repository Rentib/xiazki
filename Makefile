all: xiazki templ tailwind

xiazki:
	go build -o xiazki ./cmd/xiazki/main.go

dev:
	@make -j2 templ-dev tailwind-dev

templ-dev:
	templ generate \
		--watch \
		--proxy="http://localhost:8080" \
		--cmd="go run ./cmd/xiazki/main.go" \
		--include-version \
		--log-level=warn

tailwind-dev:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./web/static/css/tailwind.css --watch

templ:
	templ generate \
		--include-version \
		--log-level=warn

tailwind:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./web/static/css/tailwind.css

.PHONY: all templ tailwind dev templ-dev tailwind-dev
