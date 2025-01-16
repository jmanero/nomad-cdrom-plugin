# Makefile
default: build

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf build

.PHONY: build
build: build/nomad-cdrom-plugin

build/nomad-cdrom-plugin:
	mkdir -p $(@D)
	go build -o build/nomad-cdrom-plugin .
