.PHONY: test demo-down eval-skills web-embed build

test:
	GOCACHE=$${GOCACHE:-/tmp/mica-go-cache} go test ./...

demo:
	docker compose up --build

demo-down:
	docker compose down --volumes

eval-skills:
	sh ./scripts/eval-skills.sh

web-embed:
	npm --prefix web run build
	rm -rf cmd/mica/web/dist
	cp -R web/dist cmd/mica/web/

build: web-embed
	go build -o mica ./cmd/mica
