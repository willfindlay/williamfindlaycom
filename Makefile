IMAGE := williamfindlaycom
PORT  := 8080

.PHONY: build run watch stop

build:
	docker build -t $(IMAGE) .

run: build
	docker run --rm --env-file .env -p $(PORT):8080 --name $(IMAGE) $(IMAGE)

watch:
	find . -name '*.go' -o -name '*.html' -o -name '*.css' -o -name '*.js' | \
		entr -rs 'docker stop $(IMAGE) 2>/dev/null; $(MAKE) run'

stop:
	docker stop $(IMAGE) 2>/dev/null || true
