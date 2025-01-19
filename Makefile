# Makefile
VERSION ?= 0.0.0-local.0

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
export GOOS GOARCH

default: build

.PHONY: clean
clean:
	rm -rf build

.PHONY: test
test:
	go test -v ./cdrom

.PHONY: build
build: build/$(GOOS)_$(GOARCH)/nomad-device-cdrom

.PHONY: package
package: build/$(GOOS)_$(GOARCH)/nomad-device-cdrom_$(VERSION)_$(GOOS)_$(GOARCH).zip

## Run GitHub workflows locally
.PHONY: workflows
workflows:
	act --bind --container-daemon-socket - push
	act --bind --container-daemon-socket - --eventpath ./test/act-release-event.json release

build/$(GOOS)_$(GOARCH)/LICENSE:
	mkdir -p $(@D)
	cp LICENSE $@

build/$(GOOS)_$(GOARCH)/nomad-device-cdrom:
	mkdir -p $(@D)
	go build -v -ldflags "-X main.Version=$(VERSION)" -o $@ .

build/$(GOOS)_$(GOARCH)/nomad-device-cdrom_$(VERSION)_$(GOOS)_$(GOARCH).zip: build/$(GOOS)_$(GOARCH)/LICENSE build/$(GOOS)_$(GOARCH)/nomad-device-cdrom
	cd $(@D); zip $(@F) nomad-device-cdrom LICENSE
