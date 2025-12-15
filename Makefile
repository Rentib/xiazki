all: xiazki

xiazki: templ tailwind
	go build -o xiazki ./cmd/xiazki/main.go

dev:
	@make -j4 dev-server dev-templ dev-tailwind dev-sync

dev-server:
	air \
		--build.bin "./tmp/xiazki" \
		--build.cmd "go build -o ./tmp/xiazki ./cmd/xiazki/" \
		--build.delay "100" \
		--build.exclude_dir "assets,tmp,node_modules" \
		--build.exclude_regex "_test.go" \
		--build.include_ext "go" \
		--build.log "./tmp/server-build-errors.log"

dev-templ:
	templ generate \
		--watch \
		--proxy="http://localhost:8080" \
		--open-browser="false" \
		--include-version \
		--log-level=warn

dev-tailwind:
	npx @tailwindcss/cli \
		--input ./assets/css/input.css \
		--output ./web/static/css/tailwind.css \
		--watch

dev-sync:
	air \
		--build.bin "true" \
		--build.cmd "templ generate --notify-proxy" \
		--build.delay "100" \
		--build.exclude_dir "assets" \
		--build.include_ext "css,js" \
		--build.log "./tmp/sync-build-errors.log"

templ:
	templ generate \
		--include-version \
		--log-level=warn

tailwind:
	npx @tailwindcss/cli \
		--input ./assets/css/input.css \
		--output ./web/static/css/tailwind.css \
		--optimize \
		--minify

.PHONY: all templ tailwind dev templ-dev tailwind-dev
