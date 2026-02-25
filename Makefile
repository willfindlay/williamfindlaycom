IMAGE   := williamfindlaycom
PORT    := 8080
ENV_FILE := .env

.PHONY: build run watch stop

build:
	docker build -t $(IMAGE) .

run: build
	docker run --rm --env-file $(ENV_FILE) -p $(PORT):8080 --name $(IMAGE) $(IMAGE)

watch:
	find . -name '*.go' -o -name '*.html' -o -name '*.css' -o -name '*.js' | \
		entr -rs 'docker stop $(IMAGE) 2>/dev/null; $(MAKE) run'

stop:
	docker stop $(IMAGE) 2>/dev/null || true
